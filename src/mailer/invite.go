package mailer

import (
	"bytes"
	"context"
	"log"
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hadleyso/netid-activate/src/emailTemplate"
)

func HandleSendInvite(email string) error {

	// 1. Load AWS SDK configuration (uses AWS_REGION and credentials env vars)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Println("HandleSendInvite() unable to load AWS config")
		return err
	}

	// 2. Create an SESv2 client
	client := sesv2.NewFromConfig(cfg)

	// 3. Parse templates
	htmlTpl := template.Must(template.ParseFS(emailTemplate.TemplateFS, "templates/invite.html"))
	textTpl := template.Must(template.ParseFS(emailTemplate.TemplateFS, "templates/invite.txt"))
	vars := struct {
		ServiceProvider string
		PrivacyPolicy   string
		SiteName        string
		Tenant          string
	}{
		ServiceProvider: os.Getenv("LINK_SERVICE_PROVIDER"),
		PrivacyPolicy:   os.Getenv("LINK_PRIVACY_POLICY"),
		SiteName:        os.Getenv("SITE_NAME"),
		Tenant:          os.Getenv("TENANT_NAME"),
	}
	// 4. Execute into buffers
	var htmlBody, textBody bytes.Buffer
	if err := htmlTpl.Execute(&htmlBody, vars); err != nil {
		log.Println("HandleSendInvite() error render HTML template")
		return err
	}
	if err := textTpl.Execute(&textBody, vars); err != nil {
		log.Println("HandleSendInvite() error render text template")
		return err
	}

	// 5. Prepare email parameters
	from := os.Getenv("EMAIL_FROM")
	to := email
	subject := "Test SESv2 Email"

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
	resp, err := client.SendEmail(context.TODO(), input)
	if err != nil {
		log.Println("HandleSendInvite() to send email")
	}

	log.Printf("Email sent! Message ID: %s\n", *resp.MessageId)

	return nil

}
