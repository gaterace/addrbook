package addrservice

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gaterace/dml-go/pkg/dml"
	"github.com/go-kit/kit/log/level"

	_ "github.com/go-sql-driver/mysql"

	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"
)

// create a new address for a party
func (s *addrService) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.CreateAddressResponse, error) {
	resp := &pb.CreateAddressResponse{}

	// validate all inputs
	var invalidFields []string

	_, ok := addrTypeMap[req.GetAddressType()]; if !ok {
		invalidFields = append(invalidFields, "address_type")
	}

	if !isValidName(req.GetAddress_1()) {
		invalidFields = append(invalidFields, "address_1")
	}

	if (req.GetAddress_2() != "") && !isValidName(req.GetAddress_2()) {
		invalidFields = append(invalidFields, "address_2")
	}

	if !isValidCity(req.GetCity()) {
		invalidFields = append(invalidFields, "city")
	}

	if !isValidState(req.GetState()) {
		invalidFields = append(invalidFields, "state")
	}

	if !isValidPostalCode(req.GetPostalCode()) {
		invalidFields = append(invalidFields, "postal_code")
	}

	if !isValidCountryCode(req.GetCountryCode()) {
		invalidFields = append(invalidFields, "country_code")
	}

	if len(invalidFields) > 0 {
		resp.ErrorCode = 406
		resp.ErrorMessage = fmt.Sprintf("invalid fields: %s", strings.Join(invalidFields, ","))
		return resp, nil
	}

	sqlstring := `INSERT INTO tb_Address
	(inbPartyId, intAddressType, dtmCreated, dtmModified, dtmDeleted, bitIsDeleted, intVersion, inbMserviceId,
    chvAddress1, chvAddress2, chvCity, chvState, chvPostalCode, chvCountryCode) VALUES 
    (?, ?, NOW(), NOW(), NOW(), 0, 1, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	_, err = stmt.Exec(req.GetPartyId(), req.GetAddressType(), req.GetMserviceId(), req.GetAddress_1(),
		req.GetAddress_2(), req.GetCity(), req.GetState(), req.GetPostalCode(), req.GetCountryCode())

	if err == nil {
		resp.Version = 1
	} else {
		resp.ErrorCode = 501
		resp.ErrorMessage = err.Error()
		level.Error(s.logger).Log("what", "Exec", "error", err)
	}

	return resp, nil
}

// update an existing address for a party
func (s *addrService) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {
	resp := &pb.UpdateAddressResponse{}

	// validate all inputs
	var invalidFields []string

	_, ok := addrTypeMap[req.GetAddressType()]; if !ok {
		invalidFields = append(invalidFields, "address_type")
	}

	if !isValidName(req.GetAddress_1()) {
		invalidFields = append(invalidFields, "address_1")
	}

	if (req.GetAddress_2() != "") && !isValidName(req.GetAddress_2()) {
		invalidFields = append(invalidFields, "address_2")
	}

	if !isValidCity(req.GetCity()) {
		invalidFields = append(invalidFields, "city")
	}

	if !isValidState(req.GetState()) {
		invalidFields = append(invalidFields, "state")
	}

	if !isValidPostalCode(req.GetPostalCode()) {
		invalidFields = append(invalidFields, "postal_code")
	}

	if !isValidCountryCode(req.GetCountryCode()) {
		invalidFields = append(invalidFields, "country_code")
	}

	if len(invalidFields) > 0 {
		resp.ErrorCode = 406
		resp.ErrorMessage = fmt.Sprintf("invalid fields: %s", strings.Join(invalidFields, ","))
		return resp, nil
	}


	sqlstring := `UPDATE tb_Address SET dtmModified = NOW(), intVersion = ?, chvAddress1 = ?, chvAddress2 = ?,
    chvCity = ?, chvState = ?, chvPostalCode= ?, chvCountryCode = ? WHERE
    inbMserviceId = ? AND inbPartyId = ? AND intAddressType = ? AND intVersion = ? AND bitIsDeleted = 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetVersion()+1, req.GetAddress_1(), req.GetAddress_2(), req.GetCity(), req.GetState(),
		req.GetPostalCode(), req.GetCountryCode(), req.GetMserviceId(), req.GetPartyId(), req.GetAddressType(),
		req.GetVersion())

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

// delete an existing address for a party
func (s *addrService) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
	resp := &pb.DeleteAddressResponse{}

	sqlstring := `UPDATE tb_Address SET dtmModified = NOW(), intVersion = ?, bitIsDeleted = 1 WHERE 
    inbMserviceId = ? AND inbPartyId = ? AND intAddressType = ? AND intVersion = ? AND bitIsDeleted = 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetVersion()+1, req.GetMserviceId(), req.GetPartyId(), req.GetAddressType(),
		req.GetVersion())

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

// get an address for a party by id
func (s *addrService) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {
	resp := &pb.GetAddressResponse{}

	sqlstring := `SELECT inbPartyId, intAddressType, dtmCreated, dtmModified, intVersion, inbMserviceId, chvAddress1, 
    chvAddress2, chvCity, chvState, chvPostalCode, chvCountryCode FROM tb_Address WHERE
    inbMserviceId = ? AND inbPartyId = ? AND intAddressType = ? AND bitIsDeleted = 0`

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
	var addr pb.Address

	err = stmt.QueryRow(req.GetMserviceId(), req.GetPartyId(), req.GetAddressType()).Scan(&addr.PartyId,
		&addr.AddressType, &created, &modified, &addr.Version, &addr.MserviceId, &addr.Address_1,
		&addr.Address_2, &addr.City, &addr.State, &addr.PostalCode, &addr.CountryCode)

	if err == nil {
		addr.Created = dml.DateTimeFromString(created)
		addr.Modified = dml.DateTimeFromString(modified)
		addr.AddressTypeName = addrTypeMap[addr.AddressType]

		resp.ErrorCode = 0
		resp.Address = &addr
	} else if err == sql.ErrNoRows {
		resp.ErrorCode = 404
		resp.ErrorMessage = "not found"
	} else {
		level.Error(s.logger).Log("what", "QueryRow", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = err.Error()
	}

	return resp, nil
}
