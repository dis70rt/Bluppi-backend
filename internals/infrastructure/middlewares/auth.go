package middlewares

import (
    "context"
    "strings"

    "firebase.google.com/go/v4/auth"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func UnaryAuthInterceptor(authClient *auth.Client) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        
        if isPublicEndpoint(info.FullMethod) {
            return handler(ctx, req)
        }

        token, err := extractToken(ctx)
        if err != nil {
            return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token: %v", err)
        }

        decodedToken, err := authClient.VerifyIDToken(ctx, token)
        if err != nil {
            return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
        }

        // Inject user ID into context
        ctx = context.WithValue(ctx, UserIDKey, decodedToken.UID)
        return handler(ctx, req)
    }
}

func StreamAuthInterceptor(authClient *auth.Client) grpc.StreamServerInterceptor {
    return func(
        srv interface{},
        ss grpc.ServerStream,
        info *grpc.StreamServerInfo,
        handler grpc.StreamHandler,
    ) error {

        if isPublicEndpoint(info.FullMethod) {
            return handler(srv, ss)
        }

        token, err := extractToken(ss.Context())
        if err != nil {
            return status.Errorf(codes.Unauthenticated, "missing or invalid token: %v", err)
        }

        decodedToken, err := authClient.VerifyIDToken(ss.Context(), token)
        if err != nil {
            return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
        }

       
        wrappedStream := &authenticatedStream{
            ServerStream: ss,
            ctx:          context.WithValue(ss.Context(), UserIDKey, decodedToken.UID),
        }

        return handler(srv, wrappedStream)
    }
}


func extractToken(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", status.Error(codes.Unauthenticated, "no metadata found")
    }

    authHeader := md.Get("authorization")
    if len(authHeader) == 0 {
        return "", status.Error(codes.Unauthenticated, "no authorization header")
    }

    parts := strings.SplitN(authHeader[0], " ", 2)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return "", status.Error(codes.Unauthenticated, "invalid authorization format")
    }

    return parts[1], nil
}


func isPublicEndpoint(method string) bool {
    publicEndpoints := []string{
        "/users.UserService/CreateUser",
        "/users.UserService/CheckUsername",
        "/users.UserService/CheckEmail",
        "/users.UserService/GetUserByUsername",
		"/grpc.health.v1.Health/Check",
        "/grpc.health.v1.Health/Watch",
    }

    for _, endpoint := range publicEndpoints {
        if method == endpoint {
            return true
        }
    }
    return false
}


type authenticatedStream struct {
    grpc.ServerStream
    ctx context.Context
}

func (s *authenticatedStream) Context() context.Context {
    return s.ctx
}


func GetUserID(ctx context.Context) (string, error) {
    userID, ok := ctx.Value(UserIDKey).(string)
    if !ok || userID == "" {
        return "", status.Error(codes.Unauthenticated, "user not authenticated")
    }
    return userID, nil
}