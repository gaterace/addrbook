// Copyright 2020 Demian Harvill
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The addrauth package provide authorization for each gRPC method in MServiceAddrbook.
// The JWT extracted from the gRPC request context is used for each delegating method.

package addrauth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"

	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"

	"crypto/rsa"
	"io/ioutil"
)

var NotImplemented = errors.New("not implemented")

const (
	tokenExpiredMatch   = "Token is expired"
	tokenExpiredMessage = "token is expired"
)

type AddrAuth struct {
	pb.UnimplementedMServiceAddrbookServer
	logger          log.Logger
	db              *sql.DB
	rsaPSSPublicKey *rsa.PublicKey
	addrService     pb.MServiceAddrbookServer
}

// Get a new AddrAuth instance.
func NewAddrAuth(addrService pb.MServiceAddrbookServer) *AddrAuth {
	svc := AddrAuth{}
	svc.addrService = addrService
	return &svc
}

// Set the logger for the AddrAuth instance.
func (s *AddrAuth) SetLogger(logger log.Logger) {
	s.logger = logger
}

// Set the database connection for the AddrAuth instance.
func (s *AddrAuth) SetDatabaseConnection(sqlDB *sql.DB) {
	s.db = sqlDB
}

// Set the public RSA key for the AddrAuth instance, used to validate JWT.
func (s *AddrAuth) SetPublicKey(publicKeyFile string) error {
	publicKey, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		level.Error(s.logger).Log("what", "reading publicKeyFile", "error", err)
		return err
	}

	parsedKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		level.Error(s.logger).Log("what", "ParseRSAPublicKeyFromPEM", "error", err)
		return err
	}

	s.rsaPSSPublicKey = parsedKey
	return nil
}

// Bind our AddrAuth as the gRPC api server.
func (s *AddrAuth) NewApiServer(gServer *grpc.Server) error {
	if s != nil {
		pb.RegisterMServiceAddrbookServer(gServer, s)

	}
	return nil
}

// Get the JWT from the gRPC request context.
func (s *AddrAuth) GetJwtFromContext(ctx context.Context) (*map[string]interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("cannot get metadata from context")
	}

	tokens := md["token"]

	if (tokens == nil) || (len(tokens) == 0) {
		return nil, fmt.Errorf("cannot get token from context")
	}

	tokenString := tokens[0]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		method := token.Method.Alg()
		if method != "PS256" {

			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// return []byte(mySigningKey), nil
		return s.rsaPSSPublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid json web token")
	}

	claims := map[string]interface{}(token.Claims.(jwt.MapClaims))

	return &claims, nil

}

// Get the clain value as an int64.
func GetInt64FromClaims(claims *map[string]interface{}, key string) int64 {
	var val int64

	if claims != nil {
		cval := (*claims)[key]
		if fval, ok := cval.(float64); ok {
			val = int64(fval)
		}
	}

	return val
}

// Get the claim value as a string.
func GetStringFromClaims(claims *map[string]interface{}, key string) string {
	var val string

	if claims != nil {
		cval := (*claims)[key]
		if sval, ok := cval.(string); ok {
			val = sval
		}
	}

	return val
}

// create new party
func (s *AddrAuth) CreateParty(ctx context.Context, req *pb.CreatePartyRequest) (*pb.CreatePartyResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.CreatePartyResponse{}

	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if addrsvc == "addradmin" {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.CreateParty(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "CreateParty",
		"lastname", req.GetLastName(),
		"company", req.GetCompany(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// update an existing party
func (s *AddrAuth) UpdateParty(ctx context.Context, req *pb.UpdatePartyRequest) (*pb.UpdatePartyResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.UpdatePartyResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.UpdateParty(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "UpdateParty",
		"partyid", req.GetPartyId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// delete an existing party
func (s *AddrAuth) DeleteParty(ctx context.Context, req *pb.DeletePartyRequest) (*pb.DeletePartyResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.DeletePartyResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if addrsvc == "addradmin" {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.DeleteParty(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "DeleteParty",
		"partyid", req.GetPartyId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// get party by id
func (s *AddrAuth) GetParty(ctx context.Context, req *pb.GetPartyRequest) (*pb.GetPartyResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.GetPartyResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") || (addrsvc == "addrro") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.GetParty(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "GetParty",
		"partyid", req.GetPartyId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// get parties by mservice id
func (s *AddrAuth) GetParties(ctx context.Context, req *pb.GetPartiesRequest) (*pb.GetPartiesResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.GetPartiesResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") || (addrsvc == "addrro") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.GetParties(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "GetParties",
		"mservice", req.GetMserviceId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// get party wrapper by id
func (s *AddrAuth) GetPartyWrapper(ctx context.Context, req *pb.GetPartyWrapperRequest) (*pb.GetPartyWrapperResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.GetPartyWrapperResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") || (addrsvc == "addrro") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.GetPartyWrapper(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "GetPartyWrapper",
		"mservice", req.GetMserviceId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// create a new address for a party
func (s *AddrAuth) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.CreateAddressResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.CreateAddressResponse{}

	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if addrsvc == "addradmin" {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.CreateAddress(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "CreateAddress",
		"partyid", req.GetPartyId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// update an existing address for a party
func (s *AddrAuth) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.UpdateAddressResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.UpdateAddress(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "UpdateAddress",
		"partyid", req.GetPartyId(),
		"addrtype", req.GetAddressType(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// delete an existing address for a party
func (s *AddrAuth) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.DeleteAddressResponse{}

	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if addrsvc == "addradmin" {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.DeleteAddress(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "DeleteAddress",
		"partyid", req.GetPartyId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// get an address for a party by id
func (s *AddrAuth) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.GetAddressResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") || (addrsvc == "addrro") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.GetAddress(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "GetAddress",
		"mservice", req.GetMserviceId(),
		"addrtype", req.GetAddressType(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// create a new  phone
func (s *AddrAuth) CreatePhone(ctx context.Context, req *pb.CreatePhoneRequest) (*pb.CreatePhoneResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.CreatePhoneResponse{}

	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if addrsvc == "addradmin" {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.CreatePhone(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "CreatePhone",
		"partyid", req.GetPartyId(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// update an existing phone
func (s *AddrAuth) UpdatePhone(ctx context.Context, req *pb.UpdatePhoneRequest) (*pb.UpdatePhoneResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.UpdatePhoneResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.UpdatePhone(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "UpdatePhone",
		"partyid", req.GetPartyId(),
		"phonetype", req.GetPhoneType(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// delete an existing phone
func (s *AddrAuth) DeletePhone(ctx context.Context, req *pb.DeletePhoneRequest) (*pb.DeletePhoneResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.DeletePhoneResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if addrsvc == "addradmin" {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.DeletePhone(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "DeletePhone",
		"partyid", req.GetPartyId(),
		"phonetype", req.GetPhoneType(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// get a phone for a party by id
func (s *AddrAuth) GetPhone(ctx context.Context, req *pb.GetPhoneRequest) (*pb.GetPhoneResponse, error) {
	start := time.Now().UnixNano()
	resp := &pb.GetPhoneResponse{}
	resp.ErrorCode = 401
	resp.ErrorMessage = "not authorized"

	claims, err := s.GetJwtFromContext(ctx)
	if err == nil {
		addrsvc := GetStringFromClaims(claims, "addrsvc")
		if (addrsvc == "addradmin") || (addrsvc == "addrrw") {
			req.MserviceId = GetInt64FromClaims(claims, "aid")
			resp, err = s.addrService.GetPhone(ctx, req)
		}
	} else {
		if err.Error() == tokenExpiredMatch {
			resp.ErrorCode = 498
			resp.ErrorMessage = tokenExpiredMessage
		}

		err = nil
	}

	duration := time.Now().UnixNano() - start
	level.Info(s.logger).Log("endpoint", "GetPhone",
		"partyid", req.GetPartyId(),
		"phonetype", req.GetPhoneType(),
		"errcode", resp.GetErrorCode(), "duration", duration)

	return resp, err
}

// get current server version and uptime - health check
func (s *AddrAuth) GetServerVersion(ctx context.Context, req *pb.GetServerVersionRequest) (*pb.GetServerVersionResponse, error) {
	return s.addrService.GetServerVersion(ctx, req)
}
