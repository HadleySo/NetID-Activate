package mailer

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/emailTemplate"
	"github.com/spf13/viper"
)

func HandleSendOTP(email string) error {

	// Generate
	otpCode, _ := rand.Int(rand.Reader, big.NewInt(999999))

	// Save
	otpSaveErr := db.SaveOTP(email, otpCode)
	if otpSaveErr != nil {
		return otpSaveErr
	}

	sendError := sendOTPemail(email, *otpCode)
	if sendError != nil {
		return sendError
	}
	return nil
}

func sendOTPemail(email string, otpCode big.Int) error {
	// 1. Load AWS SDK configuration (uses env vars)
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Println("HandleSendInvite() unable to load AWS config")
		return err
	}

	// 2. Create an SESv2 client
	client := sesv2.NewFromConfig(cfg)

	// 3. Parse templates
	htmlTpl := template.Must(template.ParseFS(emailTemplate.TemplateFS, "templates/otp.html"))
	textTpl := template.Must(template.ParseFS(emailTemplate.TemplateFS, "templates/otp.txt"))
	vars := struct {
		Code            string
		ServiceProvider string
		PrivacyPolicy   string
		SiteName        string
		Tenant          string
	}{
		Code:            otpCode.String(),
		ServiceProvider: viper.GetString("LINK_SERVICE_PROVIDER"),
		PrivacyPolicy:   viper.GetString("LINK_PRIVACY_POLICY"),
		SiteName:        viper.GetString("SITE_NAME"),
		Tenant:          viper.GetString("TENANT_NAME"),
	}
	// 4. Execute into buffers
	var htmlBody, textBody bytes.Buffer
	if err := htmlTpl.Execute(&htmlBody, vars); err != nil {
		log.Println("sendOTPemail() error render HTML template")
		return err
	}
	if err := textTpl.Execute(&textBody, vars); err != nil {
		log.Println("sendOTPemail() error render text template")
		return err
	}
	fmt.Println(*aws.String(htmlBody.String()))

	// 5. Prepare email parameters
	from := viper.GetString("EMAIL_FROM")
	to := email
	subject := viper.GetString("TENANT_NAME") + " Activation Code"

	input := &sesv2.SendEmailInput{
		FromEmailAddress: &from,
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: &subject},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(htmlBody.String())},
					Text: &types.Content{Data: aws.String(textBody.String())},
				},
			},
		},
	}

	// 6. Send the email
	if viper.GetString("DEV") != "true" {
		resp, err := client.SendEmail(context.TODO(), input)
		if err != nil {
			log.Println("Error sendOTPemail() to send email: " + err.Error())
			return err
		}

		log.Printf("Email sent! Message ID: %s\n", *resp.MessageId)
	} else {
		fmt.Println(otpCode)
	}

	return nil
}
