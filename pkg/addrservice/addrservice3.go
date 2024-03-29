// Copyright 2020-2022 Demian Harvill
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

package addrservice

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gaterace/dml-go/pkg/dml"
	"github.com/go-kit/kit/log/level"

	_ "github.com/go-sql-driver/mysql"

	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"
)

// create a new  phone
func (s *addrService) CreatePhone(ctx context.Context, req *pb.CreatePhoneRequest) (*pb.CreatePhoneResponse, error) {
	resp := &pb.CreatePhoneResponse{}

	var invalidFields []string

	_, ok := phoneTypeMap[req.GetPhoneType()]
	if !ok {
		invalidFields = append(invalidFields, "phone_type")
	}

	if !isValidPhone(req.GetPhoneNumber()) {
		invalidFields = append(invalidFields, "phone_number")
	}

	if len(invalidFields) > 0 {
		resp.ErrorCode = 406
		resp.ErrorMessage = fmt.Sprintf("invalid fields: %s", strings.Join(invalidFields, ","))
		return resp, nil
	}

	sqlstring := `INSERT INTO tb_Phone (inbPartyId, intPhoneType, dtmCreated, dtmModified, dtmDeleted, bitIsDeleted, 
    intVersion, inbMserviceId, chvPhoneNumber) VALUES (?, ?, NOW(), NOW(), NOW(), 0, 1, ?, ?)`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	_, err = stmt.Exec(req.GetPartyId(), req.GetPhoneType(), req.GetMserviceId(), req.GetPhoneNumber())

	if err == nil {
		resp.Version = 1
	} else {
		resp.ErrorCode = 501
		resp.ErrorMessage = err.Error()
		level.Error(s.logger).Log("what", "Exec", "error", err)
	}

	return resp, nil
}

// update an existing phone
func (s *addrService) UpdatePhone(ctx context.Context, req *pb.UpdatePhoneRequest) (*pb.UpdatePhoneResponse, error) {
	resp := &pb.UpdatePhoneResponse{}

	// TODO: validate all inputs

	sqlstring := `UPDATE tb_Phone SET dtmModified = NOW(), intVersion = ?, chvPhoneNumber = ? WHERE
    inbMserviceId = ? AND inbPartyId = ? AND intPhoneType = ? AND intVersion = ? AND bitIsDeleted = 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetVersion()+1, req.GetPhoneNumber(), req.GetMserviceId(), req.GetPartyId(),
		req.GetPhoneType(), req.GetVersion())

	if err == nil {
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 1 {
			resp.Version = req.GetVersion() + 1
		} else {
			resp.ErrorCode = 404
			resp.ErrorMessage = "not found"
		}
	} else {
		resp.ErrorCode = 501
		resp.ErrorMessage = err.Error()
		level.Error(s.logger).Log("what", "Exec", "error", err)
		err = nil
	}

	return resp, nil
}

// delete an existing phone
func (s *addrService) DeletePhone(ctx context.Context, req *pb.DeletePhoneRequest) (*pb.DeletePhoneResponse, error) {
	resp := &pb.DeletePhoneResponse{}

	sqlstring := `UPDATE tb_Phone SET dtmDeleted = NOW(), intVersion = ?, bitIsDeleted = 1 WHERE
    inbMserviceId = ? AND inbPartyId = ? AND intPhoneType = ? AND intVersion = ? AND bitIsDeleted = 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetVersion()+1, req.GetMserviceId(), req.GetPartyId(),
		req.GetPhoneType(), req.GetVersion())

	if err == nil {
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 1 {
			resp.Version = req.GetVersion() + 1
		} else {
			resp.ErrorCode = 404
			resp.ErrorMessage = "not found"
		}
	} else {
		resp.ErrorCode = 501
		resp.ErrorMessage = err.Error()
		level.Error(s.logger).Log("what", "Exec", "error", err)
		err = nil
	}

	return resp, nil
}

// get a phone for a party by id
func (s *addrService) GetPhone(ctx context.Context, req *pb.GetPhoneRequest) (*pb.GetPhoneResponse, error) {
	resp := &pb.GetPhoneResponse{}

	sqlstring := `SELECT inbPartyId, intPhoneType, dtmCreated, dtmModified, intVersion, inbMserviceId, 
    chvPhoneNumber FROM tb_Phone WHERE inbMserviceId = ? AND inbPartyId = ? AND intPhoneType = ? AND bitIsDeleted = 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	var created string
	var modified string
	var phone pb.Phone

	err = stmt.QueryRow(req.GetMserviceId(), req.GetPartyId(), req.GetPhoneType()).Scan(&phone.PartyId,
		&phone.PhoneType, &created, &modified, &phone.Version, &phone.MserviceId, &phone.PhoneNumber)

	if err == nil {
		phone.Created = dml.DateTimeFromString(created)
		phone.Modified = dml.DateTimeFromString(modified)
		phone.PhoneTypeName = phoneTypeMap[phone.GetPhoneType()]
		resp.Phone = &phone
	} else if err == sql.ErrNoRows {
		resp.ErrorCode = 404
		resp.ErrorMessage = "not found"
	} else {
		level.Error(s.logger).Log("what", "QueryRow", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = err.Error()
	}

	return resp, nil

	return resp, nil
}

// get current server version and uptime - health check
func (s *addrService) GetServerVersion(ctx context.Context, req *pb.GetServerVersionRequest) (*pb.GetServerVersionResponse, error) {
	resp := &pb.GetServerVersionResponse{}

	currentSecs := time.Now().Unix()
	resp.ServerVersion = "v0.9.1"
	resp.ServerUptime = currentSecs - s.startSecs

	return resp, nil
}
