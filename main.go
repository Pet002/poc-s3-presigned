package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {
	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	client := s3.NewFromConfig(awsConfig)
	presigner := s3.NewPresignClient(client)

	presignReq, err := presigner.PresignPostObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("my-tanachod-test"),
		Key:    aws.String("/123.txt"),
	}, func(po *s3.PresignPostOptions) {
		po.Expires = time.Second * 20
	})
	if err != nil {
		panic(err)
	}
	// fmt.Println(presignReq)
	fmt.Println(presignReq.URL)
	fmt.Println(presignReq.Values)
	var bytebuffer bytes.Buffer
	writer := multipart.NewWriter(&bytebuffer)
	for key, val := range presignReq.Values {
		err := writer.WriteField(key, val)
		if err != nil {
			panic(err)
		}
	}

	writer.CreateFormFile("file", "/123.txt")

	req, err := http.NewRequest("POST", presignReq.URL, &bytebuffer)
	if err != nil {
		panic(err)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()

	httpClinet := &http.Client{
		Transport: transport,
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := httpClinet.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Status)
	b, _ := io.ReadAll(res.Body)
	fmt.Println(string(b))
}
