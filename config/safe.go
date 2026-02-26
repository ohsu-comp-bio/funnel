package config

import "google.golang.org/protobuf/proto"

// Safe returns a copy of the config with sensitive fields redacted
func (c *Config) Safe() *Config {
	if c == nil {
		return nil
	}

	cloned := proto.Clone(c)
	safe, ok := cloned.(*Config)
	if !ok {
		// Fallback: if cloning fails for some unexpected reason, avoid redacting the live config.
		copy := *c
		safe = &copy
	}

	// Database credentials
	if safe.MongoDB != nil {
		m := *safe.MongoDB
		m.Password = redact(m.Password)
		safe.MongoDB = &m
	}

	if safe.Postgres != nil {
		p := *safe.Postgres
		p.Password = redact(p.Password)
		p.AdminPassword = redact(p.AdminPassword)
		safe.Postgres = &p
	}

	if safe.Elastic != nil {
		e := *safe.Elastic
		e.Password = redact(e.Password)
		e.APIKey = redact(e.APIKey)
		e.ServiceToken = redact(e.ServiceToken)
		safe.Elastic = &e
	}

	// Cloud provider credentials
	if safe.AWSBatch != nil && safe.AWSBatch.AWSConfig != nil {
		ab := *safe.AWSBatch
		ac := *ab.AWSConfig
		ac.Key = redact(ac.Key)
		ac.Secret = redact(ac.Secret)
		ab.AWSConfig = &ac
		safe.AWSBatch = &ab
	}

	if safe.DynamoDB != nil && safe.DynamoDB.AWSConfig != nil {
		d := *safe.DynamoDB
		ac := *d.AWSConfig
		ac.Key = redact(ac.Key)
		ac.Secret = redact(ac.Secret)
		d.AWSConfig = &ac
		safe.DynamoDB = &d
	}

	if safe.AmazonS3 != nil && safe.AmazonS3.AWSConfig != nil {
		s3 := *safe.AmazonS3
		ac := *s3.AWSConfig
		ac.Key = redact(ac.Key)
		ac.Secret = redact(ac.Secret)
		s3.AWSConfig = &ac
		safe.AmazonS3 = &s3
	}

	if safe.GenericS3 != nil {
		for i, s3 := range safe.GenericS3 {
			if s3 == nil {
				continue
			}
			gs3 := *s3
			gs3.Key = redact(gs3.Key)
			gs3.Secret = redact(gs3.Secret)
			safe.GenericS3[i] = &gs3
		}
	}

	if safe.PubSub != nil {
		ps := *safe.PubSub
		ps.CredentialsFile = redact(ps.CredentialsFile)
		safe.PubSub = &ps
	}

	if safe.Datastore != nil {
		ds := *safe.Datastore
		ds.CredentialsFile = redact(ds.CredentialsFile)
		safe.Datastore = &ds
	}

	if safe.GoogleStorage != nil {
		gs := *safe.GoogleStorage
		gs.CredentialsFile = redact(gs.CredentialsFile)
		safe.GoogleStorage = &gs
	}

	// Auth credentials
	if safe.Server != nil {
		s := *safe.Server

		if s.BasicAuth != nil {
			for i, cred := range s.BasicAuth {
				if cred == nil {
					continue
				}
				bc := *cred
				bc.Password = redact(bc.Password)
				s.BasicAuth[i] = &bc
			}
		}

		if s.OidcAuth != nil {
			oa := *s.OidcAuth
			oa.ClientSecret = redact(oa.ClientSecret)
			s.OidcAuth = &oa
		}

		safe.Server = &s
	}

	if safe.RPCClient != nil && safe.RPCClient.Credential != nil {
		rpc := *safe.RPCClient
		cred := *rpc.Credential
		cred.Password = redact(cred.Password)
		rpc.Credential = &cred
		safe.RPCClient = &rpc
	}

	// Storage credentials
	if safe.Swift != nil {
		sw := *safe.Swift
		sw.Password = redact(sw.Password)
		safe.Swift = &sw
	}

	if safe.FTPStorage != nil {
		ftp := *safe.FTPStorage
		ftp.Password = redact(ftp.Password)
		safe.FTPStorage = &ftp
	}

	return safe
}

func redact(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}
