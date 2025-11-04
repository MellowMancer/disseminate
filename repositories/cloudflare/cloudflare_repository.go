package cloudflare

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "mime/multipart"
	"log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type CloudflareRepository struct {
    CloudflareAccountID       string
    CloudflareS3APIURL        string
	CloudflareToken           string
    CloudflareS3AccessKeyID   string
    CloudflareS3SecretAccessKey string
    S3Client                  *s3.Client
}

const bucketName = "mediabucket"

func NewCloudflareRepository(ctx context.Context, cloudflareAccountID, cloudflareS3APIURL, cloudflareToken, accessKeyID, secretAccessKey string) (*CloudflareRepository, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion("auto"),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),

    )
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
    }

    s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.UsePathStyle = true
		o.BaseEndpoint = aws.String(cloudflareS3APIURL)
    })

    return &CloudflareRepository{
        CloudflareAccountID:         cloudflareAccountID,
        CloudflareS3APIURL:          cloudflareS3APIURL,
		CloudflareToken:            cloudflareToken,
        CloudflareS3AccessKeyID:     accessKeyID,
        CloudflareS3SecretAccessKey: secretAccessKey,
        S3Client:                   s3Client,
    }, nil
}

func (c *CloudflareRepository) UploadFile(file multipart.File, fileName string, mimeType string) (string, error) {
    log.Println("[UPLOAD_FILE] --- Starting file upload")
    log.Printf("[UPLOAD_FILE] --- File name: %s", fileName)
    log.Printf("[UPLOAD_FILE] --- Expected MIME type: %s", mimeType)
	
    // Read the file bytes from the start.
	file.Seek(0, io.SeekStart)
    fileBytes, err := io.ReadAll(file)
    if err != nil {
        log.Printf("[UPLOAD_FILE] --- ERROR while reading file: %v", err)
        return "", fmt.Errorf("failed to read file: %w", err)
    }
    log.Printf("[UPLOAD_FILE] --- Read %d bytes from file", len(fileBytes))

    if len(fileBytes) < 10 {
        log.Printf("[UPLOAD_FILE] --- WARNING: File size looks abnormally small")
    }

    // Log the start of S3 PutObject preparation.
    log.Println("[UPLOAD_FILE] --- Preparing PutObjectInput for S3 client")
    log.Printf("[UPLOAD_FILE] --- S3 Bucket: %s", bucketName)
    log.Printf("[UPLOAD_FILE] --- S3 Key: %s", fileName)
    log.Printf("[UPLOAD_FILE] --- S3 ContentType: %s", mimeType)
    log.Printf("[UPLOAD_FILE] --- S3 ACL: %v", types.ObjectCannedACLPublicRead)

    input := &s3.PutObjectInput{
        Bucket:      aws.String(bucketName),
        Key:         aws.String(fileName),
        Body:        bytes.NewReader(fileBytes),
        ACL:         types.ObjectCannedACLPublicRead,
        ContentType: aws.String(mimeType),
    }
    log.Println("[UPLOAD_FILE] --- PutObjectInput successfully built")

    ctx := context.Background()
    log.Println("[UPLOAD_FILE] --- Calling S3Client.PutObject ...")
    result, err := c.S3Client.PutObject(ctx, input)
    if err != nil {
        log.Printf("[UPLOAD_FILE] --- ERROR during file upload: %v", err)
        return "", fmt.Errorf("failed to upload file: %w", err)
    }
    log.Println("[UPLOAD_FILE] --- File uploaded successfully")
    log.Printf("[UPLOAD_FILE] --- S3 PutObject result: %#v", result)

    publicURL := fmt.Sprintf("%s%s", "https://pub-16ef3834c60f45cca08f78c4653d8f49.r2.dev/", fileName)
    log.Printf("[UPLOAD_FILE] --- Public URL constructed: %s", publicURL)
    return publicURL, nil
}


