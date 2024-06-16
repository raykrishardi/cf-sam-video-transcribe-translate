package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"cf-sam-video-transcription-translate/config"
	osrepo "cf-sam-video-transcription-translate/pkg/repository/objectstore"
	tlrepo "cf-sam-video-transcription-translate/pkg/repository/translate"
	osuc "cf-sam-video-transcription-translate/pkg/usecase/objectstore"
	tluc "cf-sam-video-transcription-translate/pkg/usecase/translate"
	"cf-sam-video-transcription-translate/pkg/utils"

	"cf-sam-video-transcription-translate/pkg/entity"
)

var (
	AWS_REGION                       = os.Getenv("AWS_REGION")
	SOURCE_BUCKET_NAME               = os.Getenv("SOURCE_BUCKET_NAME")
	DESTINATION_BUCKET_NAME          = os.Getenv("DESTINATION_BUCKET_NAME")
	TRANSLATION_TARGET_LANGUAGE_CODE = os.Getenv("TRANSLATION_TARGET_LANGUAGE_CODE")
)

func handler(ctx context.Context, event entity.AWSEventBridgeS3Event) ([]byte, error) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("Error serializing event to JSON:%v\n", err)
	}
	log.Printf("event: %s\n", eventBytes)

	// Initialise app config
	appConfig := &config.AppConfig{
		AWSRegion:               AWS_REGION,
		TranscriptionBucketName: SOURCE_BUCKET_NAME,
		TranslationBucketName:   DESTINATION_BUCKET_NAME,
	}

	// Initialise S3 client
	s3Client, err := utils.GetAWSS3Client(ctx)
	if err != nil {
		log.Fatalf("Error getting s3 client:%v\n", err)
	}

	// Initialise repositories
	s3Repo := osrepo.NewS3Repo(appConfig, s3Client)
	tlRepo := tlrepo.NewTranslateTranscriptionRepository(appConfig)
	tlrepo.NewTranslate(tlRepo)

	// Initialise specific usecases
	osUC := osuc.NewObjectStoreUseCase(appConfig, s3Repo)
	tlUC := tluc.NewTranslateUseCase(ctx, tlRepo)

	// Initialise global usecase (if necessary)
	// uc := usecase.NewUseCase(nil, nil, tlUC, osUC)

	// Business logic
	s3GetObjectInput := entity.GetObjectInput{
		BucketName: event.Detail.Bucket.Name,
		Key:        event.Detail.Object.Key,
	}
	s3ObjectBytes, err := osUC.GetObject(ctx, s3GetObjectInput)
	if err != nil {
		log.Fatalf("Unable to get s3 content for key %s and bucket %s: %v\n", s3GetObjectInput.Key, s3GetObjectInput.BucketName, err)
	}

	// List of supported language code (https://docs.aws.amazon.com/translate/latest/dg/what-is-languages.html)
	sourceLanguageCode := "auto"
	targetLanguageCode := TRANSLATION_TARGET_LANGUAGE_CODE
	translateDocumentInput := entity.TranslateDocumentInput{
		Content:            s3ObjectBytes,
		ContentType:        "text/plain",
		SourceLanguageCode: &sourceLanguageCode,
		TargetLanguageCode: &targetLanguageCode,
	}
	translateDocumentOutput, err := tlUC.TranslateDocument(ctx, translateDocumentInput)
	if err != nil {
		log.Fatalf("Unable to translate s3 content for key %s and bucket %s: %v\n", s3GetObjectInput.Key, s3GetObjectInput.BucketName, err)
	}

	s3PutObjectInput := entity.PutObjectInput{
		BucketName: appConfig.TranslationBucketName,
		Key:        fmt.Sprintf("%s/%s", *translateDocumentInput.TargetLanguageCode, event.Detail.Object.Key),
		Body:       translateDocumentOutput.TranslatedDocument.Content,
	}
	err = osUC.PutObject(ctx, s3PutObjectInput)
	if err != nil {
		log.Fatalf("Unable to upload translated content to key %s and bucket %s: %v\n", s3PutObjectInput.Key, s3PutObjectInput.BucketName, err)
	}

	resultBytes, err := json.Marshal(translateDocumentOutput)
	if err != nil {
		log.Fatalf("Error serializing translateDocumentOutput to JSON:%v\n", err)
	}
	log.Printf("result: %s\n", resultBytes)

	return resultBytes, nil
}

func main() {
	lambda.Start(handler)
}
