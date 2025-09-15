package mailer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hadleyso/netid-activate/src/emailTemplate"
)

func HandleSendInvite(email string) error {

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
	htmlTpl := template.Must(template.ParseFS(emailTemplate.TemplateFS, "templates/invite.html"))
	textTpl := template.Must(template.ParseFS(emailTemplate.TemplateFS, "templates/invite.txt"))

	serverURL := os.Getenv("SERVER_HOSTNAME")
	if os.Getenv("OIDC_SERVER_PORT") != "" {
		serverURL = os.Getenv("SERVER_HOSTNAME") + ":" + os.Getenv("OIDC_SERVER_PORT")
	}

	vars := struct {
		ServiceProvider string
		PrivacyPolicy   string
		SiteName        string
		Tenant          string
		ServerURL       string
	}{
		ServiceProvider: os.Getenv("LINK_SERVICE_PROVIDER"),
		PrivacyPolicy:   os.Getenv("LINK_PRIVACY_POLICY"),
		SiteName:        os.Getenv("SITE_NAME"),
		Tenant:          os.Getenv("TENANT_NAME"),
		ServerURL:       serverURL,
	}
	// 4. Execute into buffers
	var htmlBody, textBody bytes.Buffer
	if err := htmlTpl.Execute(&htmlBody, vars); err != nil {
		log.Println("HandleSendInvite() error render HTML template" + err.Error())
		return err
	}
	if err := textTpl.Execute(&textBody, vars); err != nil {
		log.Println("HandleSendInvite() error render text template")
		return err
	}
	fmt.Println(*aws.String(htmlBody.String()))

	// 5. Prepare email parameters
	from := os.Getenv("EMAIL_FROM")
	to := email
	subject := os.Getenv("TENANT_NAME") + " Invite"

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
	if os.Getenv("DEV") != "true" {
		resp, err := client.SendEmail(context.TODO(), input)
		if err != nil {
			log.Println("Error HandleSendInvite() to send email" + err.Error())
			return err
		}

		log.Printf("Email sent! Message ID: %s\n", *resp.MessageId)
	}

	return nil

}
