package config

import "google.golang.org/protobuf/proto"

// Safe returns a copy of the config with sensitive fields redacted
func (c *Config) Safe() *Config {
	if c == nil {
		return nil
	}

	safe, ok := proto.Clone(c).(*Config)
	if !ok {
		copy := *c
		safe = &copy
	}

	// Database credentials
	if safe.MongoDB != nil {
		m := proto.Clone(safe.MongoDB).(*MongoDB)
		m.Password = redact(m.Password)
		safe.MongoDB = m
	}

	if safe.Postgres != nil {
		p := proto.Clone(safe.Postgres).(*Postgres)
		p.User = redact(p.User)
		p.Password = redact(p.Password)
		p.AdminUser = redact(p.AdminUser)
		p.AdminPassword = redact(p.AdminPassword)
		safe.Postgres = p
	}

	if safe.Elastic != nil {
		e := proto.Clone(safe.Elastic).(*Elastic)
		e.Password = redact(e.Password)
		e.APIKey = redact(e.APIKey)
		e.ServiceToken = redact(e.ServiceToken)
		safe.Elastic = e
	}

	// Cloud provider credentials
	if safe.AWSBatch != nil && safe.AWSBatch.AWSConfig != nil {
		ab := proto.Clone(safe.AWSBatch).(*AWSBatch)
		ac := proto.Clone(ab.AWSConfig).(*AWSConfig)
		ac.Key = redact(ac.Key)
		ac.Secret = redact(ac.Secret)
		ab.AWSConfig = ac
		safe.AWSBatch = ab
	}

	if safe.DynamoDB != nil && safe.DynamoDB.AWSConfig != nil {
		d := proto.Clone(safe.DynamoDB).(*DynamoDB)
		ac := proto.Clone(d.AWSConfig).(*AWSConfig)
		ac.Key = redact(ac.Key)
		ac.Secret = redact(ac.Secret)
		d.AWSConfig = ac
		safe.DynamoDB = d
	}

	if safe.AmazonS3 != nil && safe.AmazonS3.AWSConfig != nil {
		s3 := proto.Clone(safe.AmazonS3).(*AmazonS3Storage)
		ac := proto.Clone(s3.AWSConfig).(*AWSConfig)
		ac.Key = redact(ac.Key)
		ac.Secret = redact(ac.Secret)
		s3.AWSConfig = ac
		safe.AmazonS3 = s3
	}

	if safe.GenericS3 != nil {
		for i, s3 := range safe.GenericS3 {
			if s3 == nil {
				continue
			}
			gs3 := proto.Clone(s3).(*GenericS3Storage)
			gs3.Key = redact(gs3.Key)
			gs3.Secret = redact(gs3.Secret)
			safe.GenericS3[i] = gs3
		}
	}

	if safe.PubSub != nil {
		ps := proto.Clone(safe.PubSub).(*PubSub)
		ps.CredentialsFile = redact(ps.CredentialsFile)
		safe.PubSub = ps
	}

	if safe.Datastore != nil {
		ds := proto.Clone(safe.Datastore).(*Datastore)
		ds.CredentialsFile = redact(ds.CredentialsFile)
		safe.Datastore = ds
	}

	if safe.GoogleStorage != nil {
		gs := proto.Clone(safe.GoogleStorage).(*GoogleCloudStorage)
		gs.CredentialsFile = redact(gs.CredentialsFile)
		safe.GoogleStorage = gs
	}

	// Auth credentials
	if safe.Server != nil {
		s := proto.Clone(safe.Server).(*Server)

		if s.BasicAuth != nil {
			for i, cred := range s.BasicAuth {
				if cred == nil {
					continue
				}
				bc := proto.Clone(cred).(*BasicCredential)
				bc.Password = redact(bc.Password)
				s.BasicAuth[i] = bc
			}
		}

		if s.OidcAuth != nil {
			oa := proto.Clone(s.OidcAuth).(*OidcAuth)
			oa.ClientSecret = redact(oa.ClientSecret)
			s.OidcAuth = oa
		}

		safe.Server = s
	}

	if safe.RPCClient != nil && safe.RPCClient.Credential != nil {
		rpc := proto.Clone(safe.RPCClient).(*RPCClient)
		cred := proto.Clone(rpc.Credential).(*BasicCredential)
		cred.Password = redact(cred.Password)
		rpc.Credential = cred
		safe.RPCClient = rpc
	}

	// Storage credentials
	if safe.Swift != nil {
		sw := proto.Clone(safe.Swift).(*SwiftStorage)
		sw.Password = redact(sw.Password)
		safe.Swift = sw
	}

	if safe.FTPStorage != nil {
		ftp := proto.Clone(safe.FTPStorage).(*FTPStorage)
		ftp.Password = redact(ftp.Password)
		safe.FTPStorage = ftp
	}

	if safe.Plugins != nil && safe.Plugins.Params != nil {
		for i, param := range safe.Plugins.Params {
			if param == "" {
				continue
			}
			safe.Plugins.Params[i] = redact(param)
		}
	}

	return safe
}

func redact(str string) string {
	if str == "" {
		return ""
	}
	return "***"
}
