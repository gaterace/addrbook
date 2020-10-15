use addrbook;

DROP TABLE IF EXISTS tb_Address;

-- address book address entity
CREATE TABLE tb_Address
(

    -- party identifier
    inbPartyId BIGINT NOT NULL,
    -- type of address record, int value of AddressType
    intAddressType INT NOT NULL,
    -- creation date
    dtmCreated DATETIME NOT NULL,
    -- modification date
    dtmModified DATETIME NOT NULL,
    -- deletion date
    dtmDeleted DATETIME NOT NULL,
    -- has record been deleted?
    bitIsDeleted BOOL NOT NULL,
    -- version of this record
    intVersion INT NOT NULL,
    -- mservice account identifier
    inbMserviceId BIGINT NOT NULL,
    -- postal address line 1
    chvAddress1 VARCHAR(100) NOT NULL,
    -- postal address line 2
    chvAddress2 VARCHAR(100) NOT NULL,
    -- postal city
    chvCity VARCHAR(50) NOT NULL,
    -- postal state
    chvState VARCHAR(50) NOT NULL,
    -- postal code
    chvPostalCode VARCHAR(20) NOT NULL,
    -- country code
    chvCountryCode CHAR(2) NOT NULL,


    PRIMARY KEY (inbPartyId,intAddressType)
) ENGINE=InnoDB;

