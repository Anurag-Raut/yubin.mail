CREATE TABLE IF NOT EXISTS emails(
  id SERIAL PRIMARY KEY,
  "sentAt" DATE DEFAULT CURRENT_DATE,
  "from" TEXT REFERENCES users(id),
  "to" TEXT ,
  data TEXT,
  mailBox TEXT REFERENCES mailboxes(name)
); 

CREATE INDEX idx_emails_from ON emails("from");
