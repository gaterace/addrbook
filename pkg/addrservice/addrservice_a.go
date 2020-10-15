package addrservice

import (
	"database/sql"
	"github.com/go-kit/kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gaterace/dml-go/pkg/dml"
	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"
)

// Generic response to set specific API method response.
type genericResponse struct {
	ErrorCode    int32
	ErrorMessage string
}

func (s *addrService) GetPartyHelper(mserviceId int64, partyId int64) (*genericResponse, *pb.Party) {
	resp := &genericResponse{}

	sqlstring := `SELECT inbPartyId, dtmCreated, dtmModified, intVersion, inbMserviceId, intPartyType, chvLastName, 
	chvMiddleName, chvFirstName, chvNickname, chvCompany, chvEmail FROM tb_Party WHERE inbMserviceId = ? AND 
	inbPartyId = ? AND bitIsDeleted = 0`

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
	var party pb.Party

	err = stmt.QueryRow(partyId, mserviceId).Scan(&party.PartyId, &created, &modified, &party.Version, &party.MserviceId,
	&party.PartyType, &party.LastName, &party.MiddleName, &party.FirstName, &party.Nickname, &party.Company,
	&party.Email)

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
		resp.ErrorCode = 0
	} else if err == sql.ErrNoRows {
		resp.ErrorCode = 404
		resp.ErrorMessage = "not found"

	} else {
		level.Error(s.logger).Log("what", "QueryRow", "error", err)
		resp.ErrorCode = 500
		resp.ErrorMessage = err.Error()

	}

	return resp, &party
}

func convertPartyToWrapper(party *pb.Party) *pb.PartyWrapper {
	wrap := pb.PartyWrapper{}
	wrap.PartyId = party.GetPartyId()
	wrap.Created = party.GetCreated()
	wrap.Modified = party.GetModified()
	wrap.Version = party.GetVersion()
	wrap.MserviceId = party.GetMserviceId()
	wrap.PartyType = party.GetPartyType()
	wrap.LastName = party.GetLastName()
	wrap.MiddleName = party.GetMiddleName()
	wrap.FirstName = party.GetFirstName()
	wrap.Nickname = party.GetNickname()
	wrap.Company = party.GetCompany()
	wrap.Email = party.GetEmail()

	return &wrap
}
