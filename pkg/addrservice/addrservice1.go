package addrservice

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gaterace/dml-go/pkg/dml"
	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"

	_ "github.com/go-sql-driver/mysql"

	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"
	"google.golang.org/grpc"
)

var NotImplemented = errors.New("not implemented")

type addrService struct {
	logger    log.Logger
	db        *sql.DB
	startSecs int64
}

// Get a new addrService instance.
func NewAddrService() *addrService {
	svc := addrService{}
	svc.startSecs = time.Now().Unix()
	return &svc
}

// Set the logger for the addrService instance.
func (s *addrService) SetLogger(logger log.Logger) {
	s.logger = logger
}

// Set the database connection for the addrService instance.
func (s *addrService) SetDatabaseConnection(sqlDB *sql.DB) {
	s.db = sqlDB
}

// Bind this addrService the gRPC server api.
func (s *addrService) NewApiServer(gServer *grpc.Server) error {
	if s != nil {
		pb.RegisterMServiceAddrbookServer(gServer, s)

	}
	return nil
}

// create new party
func (s *addrService) CreateParty(ctx context.Context, req *pb.CreatePartyRequest) (*pb.CreatePartyResponse, error) {
	resp := &pb.CreatePartyResponse{}

	// TODO: validate all inputs

	sqlstring := `INSERT INTO tb_Party
      (dtmCreated, dtmModified, dtmDeleted, bitIsDeleted, intVersion, inbMserviceId, intPartyType, chvLastName,
      chvMiddleName, chvFirstName, chvNickname, chvCompany, chvEmail) 
      VALUES (NOW(), NOW(), NOW(), 0, 1, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetMserviceId(), req.GetPartyType(), req.GetLastName(), req.GetMiddleName(),
		req.GetFirstName(), req.GetNickname(), req.GetCompany(), req.GetEmail())

	if err == nil {
		partyId, err := res.LastInsertId()
		if err != nil {
			level.Error(s.logger).Log("what", "LastInsertId", "error", err)
		} else {
			level.Debug(s.logger).Log("partyId", partyId)
		}

		resp.PartyId = partyId
		resp.Version = 1
	} else {
		resp.ErrorCode = 501
		resp.ErrorMessage = err.Error()
		level.Error(s.logger).Log("what", "Exec", "error", err)
		err = nil
	}

	return resp, err
}

// update an existing party
func (s *addrService) UpdateParty(ctx context.Context, req *pb.UpdatePartyRequest) (*pb.UpdatePartyResponse, error) {
	resp := &pb.UpdatePartyResponse{}

	// TODO: validate all inputs

	sqlstring := `UPDATE tb_Party SET dtmModified = NOW(), intVersion = ?, intPartyType = ?, chvLastName = ?,
    chvMiddleName = ?, chvFirstName = ?, chvNickname = ?, chvCompany = ?, chvEmail= ? 
    WHERE inbMserviceId = ? AND inbPartyId = ? AND intVersion = ? AND  bitIsDeleted= 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetVersion()+1, req.GetPartyType(), req.GetLastName(), req.GetMiddleName(),
		req.GetFirstName(), req.GetNickname(), req.GetCompany(), req.GetEmail(), req.GetMserviceId(), req.GetPartyId(),
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

	return resp, err
}

// delete an existing party
func (s *addrService) DeleteParty(ctx context.Context, req *pb.DeletePartyRequest) (*pb.DeletePartyResponse, error) {
	resp := &pb.DeletePartyResponse{}

	sqlstring := `UPDATE tb_Party SET dtmDeleted = NOW(), intVersion = ?, bitIsDeleted = 1
    WHERE inbMserviceId = ? AND inbPartyId = ? AND intVersion = ? AND  bitIsDeleted= 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	res, err := stmt.Exec(req.GetVersion()+1, req.GetMserviceId(), req.GetPartyId(), req.GetVersion())
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

	return resp, err
}

// get party by id
func (s *addrService) GetParty(ctx context.Context, req *pb.GetPartyRequest) (*pb.GetPartyResponse, error) {
	resp := &pb.GetPartyResponse{}

	gResp, party := s.GetPartyHelper(req.GetPartyId(), req.GetMserviceId())
	resp.ErrorCode = gResp.ErrorCode
	resp.ErrorMessage = gResp.ErrorMessage
	if gResp.ErrorCode == 0 {
		resp.Party = party
	}

	return resp, nil

}

// get parties by mservice id
func (s *addrService) GetParties(ctx context.Context, req *pb.GetPartiesRequest) (*pb.GetPartiesResponse, error) {
	resp := &pb.GetPartiesResponse{}

	sqlstring := `SELECT inbPartyId, dtmCreated, dtmModified, intVersion, inbMserviceId, intPartyType, chvLastName, 
	chvMiddleName, chvFirstName, chvNickname, chvCompany, chvEmail FROM tb_Party WHERE inbMserviceId = ? AND 
	bitIsDeleted = 0`

	stmt, err := s.db.Prepare(sqlstring)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt.Close()

	rows, err := stmt.Query(req.GetMserviceId())

	if err != nil {
		level.Error(s.logger).Log("what", "Query", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = err.Error()
		return resp, nil
	}

	defer rows.Close()

	for rows.Next() {
		var created string
		var modified string
		var party pb.Party

		err = rows.Scan(&party.PartyId, &created, &modified, &party.Version, &party.MserviceId, &party.PartyType,
			&party.LastName, &party.MiddleName, &party.FirstName, &party.Nickname, &party.Company, &party.Email)

		if err == nil {
			party.Created = dml.DateTimeFromString(created)
			party.Modified = dml.DateTimeFromString(modified)
			if party.PartyType == 1 {
				party.PartyTypeName = "person"
			} else if party.PartyType == 2 {
				party.PartyTypeName = "business"
			} else {
				party.PartyTypeName = "unknown"
			}

			resp.Parties = append(resp.Parties, &party)
		} else {
			level.Error(s.logger).Log("what", "Scan", "error", err)
			resp.ErrorCode = 500
			resp.ErrorMessage = err.Error()
		}
	}

	return resp, nil

}

// get party wrapper by id
func (s *addrService) GetPartyWrapper(ctx context.Context, req *pb.GetPartyWrapperRequest) (*pb.GetPartyWrapperResponse, error) {
	resp := &pb.GetPartyWrapperResponse{}

	gResp, party := s.GetPartyHelper(req.GetPartyId(), req.GetMserviceId())
	if gResp.ErrorCode != 0 {
		resp.ErrorCode = gResp.ErrorCode
		resp.ErrorMessage = gResp.ErrorMessage
		return resp, nil
	}

	wrap := convertPartyToWrapper(party)
	sqlstring1 := `SELECT inbPartyId, intAddressType, dtmCreated, dtmModified, intVersion, inbMserviceId, chvAddress1, 
    chvAddress2, chvCity, chvState, chvPostalCode, chvCountryCode FROM tb_Address WHERE
    inbMserviceId = ? AND inbPartyId = ? AND bitIsDeleted = 0`

	stmt1, err := s.db.Prepare(sqlstring1)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt1.Close()

	rows1, err := stmt1.Query(req.GetMserviceId(), req.GetPartyId())

	if err != nil {
		level.Error(s.logger).Log("what", "Query", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = err.Error()
		return resp, nil
	}

	defer rows1.Close()

	for rows1.Next() {
		var created string
		var modified string
		var addr pb.Address

		err = rows1.Scan(&addr.PartyId, &addr.AddressType, &created, &modified, &addr.Version, &addr.MserviceId,
			&addr.Address_1, &addr.Address_2, &addr.City, &addr.State, &addr.PostalCode, &addr.CountryCode)

		if err == nil {
			addr.Created = dml.DateTimeFromString(created)
			addr.Modified = dml.DateTimeFromString(modified)
			if addr.AddressType == 1 {
				addr.AddressTypeName = "home"
			} else if addr.AddressType == 2 {
				addr.AddressTypeName = "shipping"
			} else {
				addr.AddressTypeName = "unknown"
			}

			wrap.Addresses = append(wrap.Addresses, &addr)
		} else {
			level.Error(s.logger).Log("what", "Scan", "error", err)
			resp.ErrorCode = 500
			resp.ErrorMessage = err.Error()
			return resp, nil
		}
	}

	sqlstring2 := `SELECT inbPartyId, intPhoneType, dtmCreated, dtmModified, intVersion, inbMserviceId, 
    chvPhoneNumber FROM tb_Phone WHERE inbMserviceId = ? AND inbPartyId = ? AND bitIsDeleted = 0 ORDER BY intPhoneType`

	stmt2, err := s.db.Prepare(sqlstring2)
	if err != nil {
		level.Error(s.logger).Log("what", "Prepare", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = "db.Prepare failed"
		return resp, nil
	}

	defer stmt2.Close()

	rows2, err := stmt2.Query(req.GetMserviceId(), req.GetPartyId())

	if err != nil {
		level.Error(s.logger).Log("what", "Query", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = err.Error()
		return resp, nil
	}

	defer rows2.Close()

	for rows2.Next() {
		var created string
		var modified string
		var phone pb.Phone

		err = rows2.Scan(&phone.PartyId, &phone.PhoneType, &created, &modified, &phone.Version, &phone.MserviceId,
			&phone.PhoneNumber)

		if err == nil {
			phone.Created = dml.DateTimeFromString(created)
			phone.Modified = dml.DateTimeFromString(modified)
			if phone.PhoneType == 1 {
				phone.PhoneTypeName = "home"
			} else if phone.PhoneType == 2 {
				phone.PhoneTypeName = "work"
			} else if phone.PhoneType == 3 {
				phone.PhoneTypeName = "cell"
			} else {
				phone.PhoneTypeName = "unknown"
			}
			wrap.Phones = append(wrap.Phones, &phone)
		} else {
			level.Error(s.logger).Log("what", "Scan", "error", err)
			resp.ErrorCode = 500
			resp.ErrorMessage = err.Error()
			return resp, nil
		}
	}

	resp.PartyWrapper = wrap

	return resp, nil

}