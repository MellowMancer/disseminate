package cloudflare

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "mime/multipart"

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

const bucketName = "mediaBucket"

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

func (c *CloudflareRepository) UploadFile(file multipart.File, fileName string) (string, error) {
    fileBytes, err := io.ReadAll(file)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }

    input := &s3.PutObjectInput{
        Bucket:      aws.String(bucketName),
        Key:         aws.String(fileName),
        Body:        bytes.NewReader(fileBytes),
        ACL:         types.ObjectCannedACLPublicRead, // make public, if bucket policy allows
        ContentType: aws.String("application/octet-stream"), // change as needed
    }

	ctx := context.Background()

    _, err = c.S3Client.PutObject(ctx, input)
    if err != nil {
        return "", fmt.Errorf("failed to upload file: %w", err)
    }

    // Construct the public URL
    publicURL := fmt.Sprintf("%s/%s/%s", c.CloudflareS3APIURL, bucketName, fileName)
    return publicURL, nil
}
