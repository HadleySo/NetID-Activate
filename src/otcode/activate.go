package otcode

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hadleyso/netid-activate/src/db"
)

func HandleSendOTP(email string) error {

	// Generate
	otpCode, _ := rand.Int(rand.Reader, big.NewInt(999999))

	// Save
	otpSaveErr := db.SaveOTP(email, otpCode)
	if otpSaveErr != nil {
		return otpSaveErr
	}

	// TODO: Send email
	sendOTPemail(email, *otpCode)

	return nil
}

func sendOTPemail(email string, otpCode big.Int) {
	// 1. Load AWS SDK configuration (uses AWS_REGION and credentials env vars)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Println("sendOTPemail() unable to load AWS config")
	}

	// 2. Create an SESv2 client
	client := sesv2.NewFromConfig(cfg)

	// 3. Prepare email parameters
	from := os.Getenv("EMAIL_FROM")
	to := email
	subject := "Test SESv2 Email"
	htmlBody := "<h1>Hello from SESv2!</h1><p>This is a test email.</p>"
	textBody := "Hello from SESv2!\nThis is a test email."

	input := &sesv2.SendEmailInput{
		FromEmailAddress: &from,
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: &subject,
				},
				Body: &types.Body{
					Html: &types.Content{
						Data: &htmlBody,
					},
					Text: &types.Content{
						Data: &textBody,
					},
				},
			},
		},
	}

	// 4. Send the email
	resp, err := client.SendEmail(context.TODO(), input)
	if err != nil {
		log.Println("sendOTPemail() to send email")
	}

	log.Printf("Email sent! Message ID: %s\n", *resp.MessageId)
}
