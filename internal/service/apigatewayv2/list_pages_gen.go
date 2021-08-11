// Code generated by "aws/internal/generators/listpages/main.go -function=GetApis,GetDomainNames github.com/aws/aws-sdk-go/service/apigatewayv2"; DO NOT EDIT.

package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
)

func getAPIsPages(conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApisInput, fn func(*apigatewayv2.GetApisOutput, bool) bool) error {
	return getAPIsPagesWithContext(context.Background(), conn, input, fn)
}

func getAPIsPagesWithContext(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApisInput, fn func(*apigatewayv2.GetApisOutput, bool) bool) error {
	for {
		output, err := conn.GetApisWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}

func getDomainNamesPages(conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetDomainNamesInput, fn func(*apigatewayv2.GetDomainNamesOutput, bool) bool) error {
	return getDomainNamesPagesWithContext(context.Background(), conn, input, fn)
}

func getDomainNamesPagesWithContext(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetDomainNamesInput, fn func(*apigatewayv2.GetDomainNamesOutput, bool) bool) error {
	for {
		output, err := conn.GetDomainNamesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
