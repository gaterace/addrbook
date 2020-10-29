package addrservice

import (
	"database/sql"
	"regexp"
	"github.com/go-kit/kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gaterace/dml-go/pkg/dml"
	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"
)

var validName = regexp.MustCompile("^[-A-Za-z]{1,50}$")
var validCompany = regexp.MustCompile("^[A-Za-z0-9][-A-Za-z0-9 .]{0,99}$")
var validEmail = regexp.MustCompile("^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$")
var validAddress = regexp.MustCompile("^[A-Za-z0-9][-A-Za-z0-9 .]{0,99}$")
var validCity = regexp.MustCompile("^[A-Za-z][-A-Za-z .]{0,49}$")
var validState = regexp.MustCompile("^[-A-Za-z][-A-Za-z] {0,49}$")
var validUSZipcode = regexp.MustCompile("^([0-9]{5})([\\-]{1}[0-9]{4})?$")
var validPostalCode = regexp.MustCompile("^[-a-zA-z0-9]{5,20}$")
var validCountryCode = regexp.MustCompile("^[a-z][a-z]$")
var validPhone = regexp.MustCompile("^(\\+[1-9][0-9]{0,3}-)?[1-9][0-9]{2}-[1-9][0-9]{2}-[0-9]{4}(x[0-9]+)?$")

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
		party.PartyTypeName = partyTypeMap[party.PartyType]
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

func isValidName(name string) bool {
	return validName.MatchString(name)
}

func isValidCompany(name string) bool {
	// make sure no trailing space, regex takes care of rest
	if len(name) > 0 {
		if name[len(name) - 1] == ' ' {
			return false
		}
	}
	return validCompany.MatchString(name)
}

func isValidEmail(name string) bool {
	return validEmail.MatchString(name)
}

func isValidAddress(name string) bool {
	return validAddress.MatchString(name)
}

func isValidCity(name string) bool {
	return validCity.MatchString(name)
}

func isValidState(name string) bool {
	return validState.MatchString(name)
}

func isValidPostalCode(name string, country string) bool {
	// initial support for us zipcodes
	if country == "us" {
		return validUSZipcode.MatchString(name)
	} else {
		return validPostalCode.MatchString(name)
	}
}

func isValidCountryCode(name string) bool {
	// initial support for us only
	return validCountryCode.MatchString(name)
}

func isValidPhone(phone string) bool {
	return validPhone.MatchString(phone)
}