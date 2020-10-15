use addrbook;

DROP TABLE IF EXISTS tb_Phone;

-- address book phone entity
CREATE TABLE tb_Phone
(

    -- party identifier
    inbPartyId BIGINT NOT NULL,
    -- type of phone record, int value of PhoneType
    intPhoneType INT NOT NULL,
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
    -- phone number
    chvPhoneNumber VARCHAR(20) NOT NULL,


    PRIMARY KEY (inbPartyId,intPhoneType)
) ENGINE=InnoDB;

