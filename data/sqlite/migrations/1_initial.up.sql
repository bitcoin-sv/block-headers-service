CREATE TABLE blockheaders(
    hash VARCHAR PRIMARY KEY
    ,confirmations INTEGER
    ,height INTEGER
    ,version INTEGER
    ,versionHex VARCHAR
    ,merkleroot VARCHAR
    ,time INTEGER
    ,mediantime INTEGER
    ,nonce INTEGER
    ,bits VARCHAR
    ,difficulty FLOAT
    ,chainwork VARCHAR
    ,previousblockhash VARCHAR
    ,nextblockhash VARCHAR
    ,createdAt      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);