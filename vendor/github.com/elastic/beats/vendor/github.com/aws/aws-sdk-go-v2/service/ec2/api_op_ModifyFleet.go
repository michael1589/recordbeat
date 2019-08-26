// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/internal/awsutil"
)

// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/ModifyFleetRequest
type ModifyFleetInput struct {
	_ struct{} `type:"structure"`

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have
	// the required permissions, the error response is DryRunOperation. Otherwise,
	// it is UnauthorizedOperation.
	DryRun *bool `type:"boolean"`

	// Indicates whether running instances should be terminated if the total target
	// capacity of the EC2 Fleet is decreased below the current size of the EC2
	// Fleet.
	ExcessCapacityTerminationPolicy FleetExcessCapacityTerminationPolicy `type:"string" enum:"true"`

	// The ID of the EC2 Fleet.
	//
	// FleetId is a required field
	FleetId *string `type:"string" required:"true"`

	// The size of the EC2 Fleet.
	//
	// TargetCapacitySpecification is a required field
	TargetCapacitySpecification *TargetCapacitySpecificationRequest `type:"structure" required:"true"`
}

// String returns the string representation
func (s ModifyFleetInput) String() string {
	return awsutil.Prettify(s)
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *ModifyFleetInput) Validate() error {
	invalidParams := aws.ErrInvalidParams{Context: "ModifyFleetInput"}

	if s.FleetId == nil {
		invalidParams.Add(aws.NewErrParamRequired("FleetId"))
	}

	if s.TargetCapacitySpecification == nil {
		invalidParams.Add(aws.NewErrParamRequired("TargetCapacitySpecification"))
	}
	if s.TargetCapacitySpecification != nil {
		if err := s.TargetCapacitySpecification.Validate(); err != nil {
			invalidParams.AddNested("TargetCapacitySpecification", err.(aws.ErrInvalidParams))
		}
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/ModifyFleetResult
type ModifyFleetOutput struct {
	_ struct{} `type:"structure"`

	// Is true if the request succeeds, and an error otherwise.
	Return *bool `locationName:"return" type:"boolean"`
}

// String returns the string representation
func (s ModifyFleetOutput) String() string {
	return awsutil.Prettify(s)
}

const opModifyFleet = "ModifyFleet"

// ModifyFleetRequest returns a request value for making API operation for
// Amazon Elastic Compute Cloud.
//
// Modifies the specified EC2 Fleet.
//
// While the EC2 Fleet is being modified, it is in the modifying state.
//
//    // Example sending a request using ModifyFleetRequest.
//    req := client.ModifyFleetRequest(params)
//    resp, err := req.Send(context.TODO())
//    if err == nil {
//        fmt.Println(resp)
//    }
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/ModifyFleet
func (c *Client) ModifyFleetRequest(input *ModifyFleetInput) ModifyFleetRequest {
	op := &aws.Operation{
		Name:       opModifyFleet,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ModifyFleetInput{}
	}

	req := c.newRequest(op, input, &ModifyFleetOutput{})
	return ModifyFleetRequest{Request: req, Input: input, Copy: c.ModifyFleetRequest}
}

// ModifyFleetRequest is the request type for the
// ModifyFleet API operation.
type ModifyFleetRequest struct {
	*aws.Request
	Input *ModifyFleetInput
	Copy  func(*ModifyFleetInput) ModifyFleetRequest
}

// Send marshals and sends the ModifyFleet API request.
func (r ModifyFleetRequest) Send(ctx context.Context) (*ModifyFleetResponse, error) {
	r.Request.SetContext(ctx)
	err := r.Request.Send()
	if err != nil {
		return nil, err
	}

	resp := &ModifyFleetResponse{
		ModifyFleetOutput: r.Request.Data.(*ModifyFleetOutput),
		response:          &aws.Response{Request: r.Request},
	}

	return resp, nil
}

// ModifyFleetResponse is the response type for the
// ModifyFleet API operation.
type ModifyFleetResponse struct {
	*ModifyFleetOutput

	response *aws.Response
}

// SDKResponseMetdata returns the response metadata for the
// ModifyFleet request.
func (r *ModifyFleetResponse) SDKResponseMetdata() *aws.Response {
	return r.response
}
