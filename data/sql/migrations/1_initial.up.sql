CREATE TABLE headers(
    hash VARCHAR(255) PRIMARY KEY
    ,height INTEGER
    ,version INTEGER
    ,merkleroot VARCHAR(255)
    ,nonce BIGINT
    ,bits VARCHAR(255)
    ,chainwork VARCHAR(255)
    ,previousblock VARCHAR(255)
    ,timestamp      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,isorphan BOOLEAN
    ,isconfirmed BOOLEAN
    ,cumulatedWork VARCHAR(255)
);
