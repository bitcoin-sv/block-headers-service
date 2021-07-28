CREATE TABLE blockheaders(
    hash VARCHAR(255) PRIMARY KEY
    ,confirmations INTEGER
    ,height INTEGER
    ,version INTEGER
    ,versionhex VARCHAR(255)
    ,merkleroot VARCHAR(255)
    ,time INTEGER
    ,mediantime INTEGER
    ,nonce BIGINT
    ,bits VARCHAR(255)
    ,difficulty FLOAT
    ,chainwork VARCHAR(255)
    ,previousblockhash VARCHAR(255)
    ,nextblockhash VARCHAR(255)
    ,createdAt      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);