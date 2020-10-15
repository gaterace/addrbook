package main

import (
	"context"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strconv"

	pb "github.com/gaterace/addrbook/pkg/mserviceaddrbook"
	"github.com/kylelemons/go-gypsy/yaml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"

	flag "github.com/juju/gnuflag"
)

var id = flag.Int64("id", 0, "entity id")
var version = flag.Int("version", -1, "version")

var ptype = flag.String("ptype", "", "party type")
var lname = flag.String("lname", "", "last name")
var mname = flag.String("mname", "", "middle name")
var fname = flag.String("fname", "", "first name")
var nickname = flag.String("nickname", "", "nickname")
var company = flag.String("company", "", "company")
var email = flag.String("e", "", "email")

var atype = flag.String("atype", "", "address type")
var address_1 = flag.String("address_1", "", "address line 1")
var address_2 = flag.String("address_2", "", "address line 2")
var city = flag.String("city", "", "city")
var state = flag.String("state", "", "state")
var postal_code = flag.String("postal_code", "", "postal code")
var country_code = flag.String("country_code", "us", "country code")

var phtype = flag.String("phtype", "", "phone type")
var phone = flag.String("phone", "", "phone number")

func main() {
	flag.Parse(true)

	configFilename := "conf.yaml"
	usr, err := user.Current()
	if err == nil {
		homeDir := usr.HomeDir
		configFilename = homeDir + string(os.PathSeparator) + ".addrbook.config"
		// _ = homeDir + string(os.PathSeparator) + ".addrbook.config"
	}

	config, err := yaml.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("configuration not found: " + configFilename)
	}

	ca_file, _ := config.Get("ca_file")
	tls, _ := config.GetBool("tls")
	server_host_override, _ := config.Get("server_host_override")
	server, _ := config.Get("server")
	port, _ := config.GetInt("port")

	if port == 0 {
		port = 50057
	}

	if len(flag.Args()) < 1 {
		prog := os.Args[0]
		fmt.Printf("Command line client for addrbook grpc service\n")
		fmt.Printf("usage:\n")
		fmt.Printf("    %s create_party --ptype <party type> --fname <first name> --mname <middle name>  --lname <last name> \n", prog)
		fmt.Printf("          --nickname <nickname> --company <company> -e <email>\n")
		fmt.Printf("    %s update_party --id <party id>  --version <version> --ptype <party type>  --fname <first name>\n" , prog)
		fmt.Printf("          --mname <middle name>  --lname <last name> --nickname <nickname> --company <company> -e <email>\n")
		fmt.Printf("    %s delete_party --id <party id> --version <version>\n" , prog)
		fmt.Printf("    %s get_party --id <party id> \n" , prog)
		fmt.Printf("    %s get_parties  \n" , prog)
		fmt.Printf("    %s get_party_wrapper --id <party id> \n" , prog)
		fmt.Printf("    %s create_address --id <party id> --atype <address type> --address1 <address 1> [--address2 <address 2>]\n", prog)
		fmt.Printf("          --city <city> --state <state> --postal_code <postal code> [--country_code <country code>]\n")
		fmt.Printf("    %s update_address --id <party id> --atype <address type> --version <version> --address1 <address 1> [--address2 <address 2>]\n", prog)
		fmt.Printf("          --city <city> --state <state> --postal_code <postal code> [--country_code <country code>]\n")
		fmt.Printf("    %s delete_address --id <party id> --atype <address type> --version <version>\n", prog)
		fmt.Printf("    %s get_address --id <party id> --atype <address type> \n", prog)
		fmt.Printf("    %s create_phone --id <party id> --phtype <phone type> --phone <phone number> \n", prog)
		fmt.Printf("    %s update_phone --id <party id> --phtype <phone type> --version <version> --phone <phone number> \n", prog)
		fmt.Printf("    %s delete_phone --id <party id> --phtype <phone type> --version <version> \n", prog)
		fmt.Printf("    %s get_phone --id <party id> --phtype <phone type>  \n", prog)

		fmt.Printf("    %s get_server_version\n", prog)

		os.Exit(1)
	}

	cmd := flag.Arg(0)

	validParams := true

	switch cmd {
	case "create_party":
		if (*ptype != "person") && (*ptype != "business") {
			fmt.Println("ptype parameter missing, must be person or business")
			validParams = false
		}
		if *ptype == "business" {
			if *company == "" {
				fmt.Println("company parameter missing")
				validParams = false
			}
		}
		if *ptype == "person" {
			if *lname == "" {
				fmt.Println("lname parameter missing")
				validParams = false
			}
			if *fname == "" {
				fmt.Println("fname parameter missing")
				validParams = false
			}
		}

		if *email == "" {
			fmt.Println("email parameter missing")
			validParams = false
		}

	case "update_party":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}

		if *version < 0 {
			fmt.Println("version parameter missing")
			validParams = false
		}
		if (*ptype != "person") && (*ptype != "business") {
			fmt.Println("ptype parameter missing, must be person or business")
			validParams = false
		}
		if *ptype == "business" {
			if *company == "" {
				fmt.Println("company parameter missing")
				validParams = false
			}
		}
		if *ptype == "person" {
			if *lname == "" {
				fmt.Println("lname parameter missing")
				validParams = false
			}
			if *fname == "" {
				fmt.Println("fname parameter missing")
				validParams = false
			}
		}

		if *email == "" {
			fmt.Println("email parameter missing")
			validParams = false
		}

	case "delete_party":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}

		if *version < 0 {
			fmt.Println("version parameter missing")
			validParams = false
		}
	case "get_party":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
	case "get_parties":
		// no parameters
		validParams = true
	case "get_party_wrapper":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
	case "create_address":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*atype != "home") && (*atype != "shipping") {
			fmt.Println("atype parameter missing, must be home or shipping")
			validParams = false
		}
		if *address_1 == "" {
			fmt.Println("address_1 parameter missing")
			validParams = false
		}
		if *city == "" {
			fmt.Println("city parameter missing")
			validParams = false
		}
		if *state == "" {
			fmt.Println("state parameter missing")
			validParams = false
		}
		if *postal_code == "" {
			fmt.Println("postal_code parameter missing")
			validParams = false
		}
		if len(*country_code) != 2 {
			fmt.Println("country_code parameter must be 2 character country code")
			validParams = false
		}
	case "update_address":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*atype != "home") && (*atype != "shipping") {
			fmt.Println("atype parameter missing, must be home or shipping")
			validParams = false
		}
		if *version < 0 {
			fmt.Println("version parameter missing")
			validParams = false
		}
		if *address_1 == "" {
			fmt.Println("address_1 parameter missing")
			validParams = false
		}
		if *city == "" {
			fmt.Println("city parameter missing")
			validParams = false
		}
		if *state == "" {
			fmt.Println("state parameter missing")
			validParams = false
		}
		if *postal_code == "" {
			fmt.Println("postal_code parameter missing")
			validParams = false
		}
		if len(*country_code) != 2 {
			fmt.Println("country_code parameter must be 2 character country code")
			validParams = false
		}
	case "delete_address":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*atype != "home") && (*atype != "shipping") {
			fmt.Println("atype parameter missing, must be home or shipping")
			validParams = false
		}
		if *version < 0 {
			fmt.Println("version parameter missing")
			validParams = false
		}
	case "get_address":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*atype != "home") && (*atype != "shipping") {
			fmt.Println("atype parameter missing, must be home or shipping")
			validParams = false
		}
	case "create_phone":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*phtype != "home") && (*phtype != "work") && (*phtype != "cell") {
			fmt.Println("phtype parameter missing, must be home, work or cell")
			validParams = false
		}
		if *phone == "" {
			fmt.Println("phone parameter missing")
			validParams = false
		}
	case "update_phone":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*phtype != "home") && (*phtype != "work") && (*phtype != "cell") {
			fmt.Println("phtype parameter missing, must be home, work or cell")
			validParams = false
		}
		if *version < 0 {
			fmt.Println("version parameter missing")
			validParams = false
		}
		if *phone == "" {
			fmt.Println("phone parameter missing")
			validParams = false
		}
	case "delete_phone":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*phtype != "home") && (*phtype != "work") && (*phtype != "cell") {
			fmt.Println("phtype parameter missing, must be home, work or cell")
			validParams = false
		}
		if *version < 0 {
			fmt.Println("version parameter missing")
			validParams = false
		}
	case "get_phone":
		if *id <= 0 {
			fmt.Println("id parameter missing")
			validParams = false
		}
		if (*phtype != "home") && (*phtype != "work") && (*phtype != "cell") {
			fmt.Println("phtype parameter missing, must be home, work or cell")
			validParams = false
		}
	case "get_server_version":
		validParams = true

	default:
		fmt.Printf("unknown command: %s\n", cmd)
		validParams = false
	}

	if !validParams {
		os.Exit(1)
	}

	tokenFilename := "token.txt"
	usr, err = user.Current()
	if err == nil {
		homeDir := usr.HomeDir
		tokenFilename = homeDir + string(os.PathSeparator) + ".mservice.token"
	}

	address := server + ":" + strconv.Itoa(int(port))
	// fmt.Printf("address: %s\n", address)

	var opts []grpc.DialOption
	if tls {
		var sn string
		if server_host_override != "" {
			sn = server_host_override
		}
		var creds credentials.TransportCredentials
		if ca_file != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(ca_file, sn)
			if err != nil {
				grpclog.Fatalf("Failed to create TLS credentials %v", err)
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// set up connection to server
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewMServiceAddrbookClient(conn)
	ctx := context.Background()

	savedToken := ""

	data, err := ioutil.ReadFile(tokenFilename)

	if err == nil {
		savedToken = string(data)
	}

	md := metadata.Pairs("token", savedToken)
	mctx := metadata.NewOutgoingContext(ctx, md)

	switch cmd {
	case "create_party":
		req := pb.CreatePartyRequest{}
		var party_type int32
		if *ptype == "person" {
			party_type = 1
		} else {
			party_type = 2
		}
		req.PartyType = party_type
		req.FirstName = *fname
		req.MiddleName = *mname
		req.LastName = *lname
		req.Nickname = *nickname
		req.Company = *company
		req.Email = *email
		resp, err := client.CreateParty(mctx, &req)
		printResponse(resp, err)
	case "update_party":
		req := pb.UpdatePartyRequest{}
		var party_type int32
		if *ptype == "person" {
			party_type = 1
		} else {
			party_type = 2
		}
		req.PartyId = *id
		req.Version = int32(*version)
		req.PartyType = party_type
		req.FirstName = *fname
		req.MiddleName = *mname
		req.LastName = *lname
		req.Nickname = *nickname
		req.Company = *company
		req.Email = *email
		resp, err := client.UpdateParty(mctx, &req)
		printResponse(resp, err)

	case "get_server_version":
		req := pb.GetServerVersionRequest{}
		req.DummyParam = 1
		resp, err := client.GetServerVersion(mctx, &req)
		printResponse(resp, err)
	case "delete_party":
		req := pb.DeletePartyRequest{}
		req.PartyId = *id
		req.Version = int32(*version)
		resp, err := client.DeleteParty(mctx, &req)
		printResponse(resp, err)
	case "get_party":
		req := pb.GetPartyRequest{}
		req.PartyId = *id
		resp, err := client.GetParty(mctx, &req)
		printResponse(resp, err)
	case "get_parties":
		req := pb.GetPartiesRequest{}
		resp, err := client.GetParties(mctx, &req)
		printResponse(resp, err)
	case "get_party_wrapper":
		req := pb.GetPartyWrapperRequest{}
		req.PartyId = *id
		resp, err := client.GetPartyWrapper(mctx, &req)
		printResponse(resp, err)
	case "create_address":
		req := pb.CreateAddressRequest{}
		req.PartyId = *id
		if *atype == "home" {
			req.AddressType = 1
		} else if *atype == "shipping" {
			req.AddressType = 2
		}
		req.Address_1 = *address_1
		req.Address_2 = *address_2
		req.City = *city
		req.State = *state
		req.PostalCode = *postal_code
		req.CountryCode = *country_code
		resp, err := client.CreateAddress(mctx, &req)
		printResponse(resp, err)
	case "update_address":
		req := pb.UpdateAddressRequest{}
		req.PartyId = *id
		if *atype == "home" {
			req.AddressType = 1
		} else if *atype == "shipping" {
			req.AddressType = 2
		}
		req.Version = int32(*version)
		req.Address_1 = *address_1
		req.Address_2 = *address_2
		req.City = *city
		req.State = *state
		req.PostalCode = *postal_code
		req.CountryCode = *country_code
		resp, err := client.UpdateAddress(mctx, &req)
		printResponse(resp, err)
	case "delete_address":
		req := pb.DeleteAddressRequest{}
		req.PartyId = *id
		if *atype == "home" {
			req.AddressType = 1
		} else if *atype == "shipping" {
			req.AddressType = 2
		}
		req.Version = int32(*version)
		resp, err := client.DeleteAddress(mctx, &req)
		printResponse(resp, err)
	case "get_address":
		req := pb.GetAddressRequest{}
		req.PartyId = *id
		if *atype == "home" {
			req.AddressType = 1
		} else if *atype == "shipping" {
			req.AddressType = 2
		}
		resp, err := client.GetAddress(mctx, &req)
		printResponse(resp, err)
	case "create_phone":
		req := pb.CreatePhoneRequest{}
		req.PartyId = *id
		if *phtype == "home" {
			req.PhoneType = 1
		} else if *phtype == "work" {
			req.PhoneType = 2
		} else if *phtype == "cell" {
			req.PhoneType = 3
		}
		req.PhoneNumber = *phone
		resp, err := client.CreatePhone(mctx, &req)
		printResponse(resp, err)
	case "update_phone":
		req := pb.UpdatePhoneRequest{}
		req.PartyId = *id
		if *phtype == "home" {
			req.PhoneType = 1
		} else if *phtype == "work" {
			req.PhoneType = 2
		} else if *phtype == "cell" {
			req.PhoneType = 3
		}
		req.PhoneNumber = *phone
		req.Version = int32(*version)
		resp, err := client.UpdatePhone(mctx, &req)
		printResponse(resp, err)
	case "delete_phone":
		req := pb.DeletePhoneRequest{}
		req.PartyId = *id
		if *phtype == "home" {
			req.PhoneType = 1
		} else if *phtype == "work" {
			req.PhoneType = 2
		} else if *phtype == "cell" {
			req.PhoneType = 3
		}
		req.Version = int32(*version)
		resp, err := client.DeletePhone(mctx, &req)
		printResponse(resp, err)
	case "get_phone":
		req := pb.GetPhoneRequest{}
		req.PartyId = *id
		if *phtype == "home" {
			req.PhoneType = 1
		} else if *phtype == "work" {
			req.PhoneType = 2
		} else if *phtype == "cell" {
			req.PhoneType = 3
		}
		resp, err := client.GetPhone(mctx, &req)
		printResponse(resp, err)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

// Helper to print api method response as JSON.
func printResponse(resp interface{}, err error) {
	if err == nil {
		jtext, err := json.MarshalIndent(resp, "", "  ")
		if err == nil {
			fmt.Println(string(jtext))
		}
	}
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
}
