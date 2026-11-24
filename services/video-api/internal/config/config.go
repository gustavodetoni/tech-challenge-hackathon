package config

import (
 "os"
 "strconv"
 "strings"
 "time"
)

type Config struct { HTTPAddr, DatabaseURL, IssuerURL, JWKSURL, Audience, CORSOrigin, MinioInternalEndpoint, MinioPublicEndpoint, MinioAccessKey, MinioSecretKey, InputBucket, ResultBucket string; MaxUploadBytes int64; PresignTTL time.Duration }
func Load() Config { return Config{
 HTTPAddr: value("HTTP_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"), IssuerURL: value("KEYCLOAK_ISSUER_URL", "http://localhost:8083/realms/fiap-video"), JWKSURL: value("KEYCLOAK_JWKS_URL", "http://keycloak:8080/realms/fiap-video/protocol/openid-connect/certs"), Audience: value("KEYCLOAK_AUDIENCE", "fiap-video-spa"), CORSOrigin: value("CORS_ORIGIN", "http://localhost:5173"), MinioInternalEndpoint: value("MINIO_INTERNAL_ENDPOINT", "minio:9000"), MinioPublicEndpoint: value("MINIO_PUBLIC_ENDPOINT", "localhost:9000"), MinioAccessKey: os.Getenv("MINIO_ROOT_USER"), MinioSecretKey: os.Getenv("MINIO_ROOT_PASSWORD"), InputBucket: value("MINIO_INPUT_BUCKET", "video-inputs"), ResultBucket: value("MINIO_RESULT_BUCKET", "video-results"), MaxUploadBytes: intValue("MAX_UPLOAD_BYTES", 524288000), PresignTTL: duration("UPLOAD_URL_TTL", 15*time.Minute),
 } }
func value(k, d string) string { if v:=os.Getenv(k);v!="" { return v }; return d }
func intValue(k string,d int64) int64 { if v,e:=strconv.ParseInt(os.Getenv(k),10,64);e==nil&&v>0{return v};return d }
func duration(k string,d time.Duration) time.Duration { if v,e:=time.ParseDuration(os.Getenv(k));e==nil&&v>0{return v};return d }
func AllowedOrigins(v string) []string { return strings.FieldsFunc(v, func(r rune)bool{return r==','}) }
