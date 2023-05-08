CREATE TABLE webhooks(
    name                VARCHAR(255) PRIMARY KEY
    ,url                VARCHAR(255) UNIQUE
    ,tokenHeader        VARCHAR(255)
    ,token              VARCHAR(255)
    ,createdAt          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,lastEmitStatus     VARCHAR(255) DEFAULT ""
    ,lastEmitTimestamp  TIMESTAMP DEFAULT 0
    ,errorsCount        INTEGER DEFAULT 0
);
