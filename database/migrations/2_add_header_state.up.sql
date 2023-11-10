alter table headers add column header_state VARCHAR(50) default 'LONGEST_CHAIN';

update headers
set header_state = 'ORPHAN'
where isorphan = true;

alter table headers drop column isorphan;
alter table headers drop column isconfirmed;
