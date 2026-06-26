package acl

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AclData map[string][]string

type AclService struct {
	data AclData
}

func NewAclService(data AclData) *AclService {
	return &AclService{
		data,
	}
}

func (acl *AclService) AclInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "can't read from metadata")
	}

	consumerNames, ok := md["consumer"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata doesn't contain consumer name")
	}

	for _, method := range acl.data[consumerNames[0]] {
		if method == info.FullMethod || strings.HasSuffix(method, extractServiceFromMethod(info.FullMethod)+"/*") {
			return handler(ctx, req)
		}
	}

	return nil, status.Error(codes.Unauthenticated, "this method is forbidden")
}

func (acl *AclService) AclStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.InvalidArgument, "can't read from metadata")
	}

	consumerNames, ok := md["consumer"]
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata doesn't contain consumer name")
	}

	for _, aclRecord := range acl.data[consumerNames[0]] {
		if aclRecord == info.FullMethod || strings.HasSuffix(aclRecord, extractServiceFromMethod(info.FullMethod)+"/*") {
			return handler(srv, ss)
		}
	}

	return status.Error(codes.Unauthenticated, "this method is forbidden")
}

var serviceRexp = regexp.MustCompile(`\.([a-zA-Z]+)\/`)

func extractServiceFromMethod(method string) string {
	matches := serviceRexp.FindStringSubmatch(method)
	if len(matches) < 2 {
		return ""
	}

	fmt.Println("mathces", matches)

	return matches[1]
}
