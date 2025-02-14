---
title: SDA Storage
menu:
  main:
    parent: Storage
---

# Sensitive Data Archive (SDA) Storage

Funnel supports content-retrieval from an [SDA-Download][sda]-compatible API.
This is a service with an HTTP-based REST-API with some additions:

1. The request must be authenticated using a Bearer token, a JSON Web Token to
   be validated by SDA, checking it would contain a valid GA4GH Visa permitting
   access to the targeted dataset.
2. If the targeted file is encrypted using [Crypt4gh](crypt4gh) (it has ".c4gh"
   extension), the client (e.g. Funnel) needs to send its public key so that
   SDA would reencrypt the file header, which would enable to obtain the cipher
   key from the header using its private key. Funnel makes the file accessible
   to the computation task without encryption.

The task input file URL needs to specify `sda` as the resource protocol.
Funnel will extract path information from the specified URL and append it to
the service URL specified in the Funnel configuration. The format of the input
data URL is following:

```
sda://<dataset-id>/<resource/path>
```

For example: `sda://DATASET_2000/synthetic/sample.bam`

If the service expects a `Bearer` (token) or `Basic` (username:password)
authentication, it can be specified at the end of the URL right after the
hash-sign (`#`). For example: `sda://dataset/file#jwt-token-here`.
Note that when the task is submitted to Funnel using a valid `Bearer` token for
user authentication, the same token will be automatically appended to the
SDA URL, so the request to the SDA service would use the same token.
Exception is when the URL already specifies the hash-sign (`#`) – then the
provided value won't be replaced.

Funnel sends its Crypt4gh public key in the header (`client-public-key`) of the
request to the SDA service, when the requested file has ".c4gh" extension.

For sensitive data, the deployment environment (server) should pay attention to
restricting access to the Funnel's data directories, possibly having separate
Funnel instances for different data-projects.

SDA Storage configuration just requires a service URL to become active:

```yaml
SDAStorage:
  ServiceURL: https://example.org:8443/sda/
  Timeout: 30s
```

If the `ServiceUrl` is undefined, `sda` protocol will be disabled.

Funnel will automatically append `/s3/<provided-path>` or
`/s3-encrypted/<provided-path>` to the service URL, depending on whether the
requested file has the Crypt4GH file-extension (`.c4gh`).

### About Crypt4GH Keys

Funnel loads Crypt4gh keys from files, or generates and saves them when the
files cannot be resolved.

First, Funnel tries to resolve the public and secret key file-paths from
environment variables:

- `C4GH_SECRET_KEY` – path to the secret/private key
- `C4GH_PUBLIC_KEY` (optional) – path to public key,
- `C4GH_PASSPHRASE` (optional) – password of the secret/private key.

Notes:

- If `C4GH_PUBLIC_KEY` is provided and the file exists, it must
  cryptographically pair with the secret key.
- If `C4GH_SECRET_KEY` refers to an unencrypted secret key,`C4GH_PASSPHRASE`
  may be omitted.
- When the files of `C4GH_PUBLIC_KEY` and `C4GH_SECRET_KEY` do not exist yet,
  a new key-pair will be generated and stored in the specified files (secret
  key will be encrypted with `C4GH_SECRET_KEY`, if present).

When the variables are not declared, the local and home directory files will be
tried instead: `.c4gh/key[.pub]` and `~/.c4gh/key[.pub]` (the secret key file
here is expected to be just `key`, and public key in `key.pub`). If these files
(especially the secret key) do not exist, a new key-pair will be generated
and stored in the **home-directory** file-paths, and, on failure, in the
**local directory** file-paths.

### Example task

```json
{
  "name": "Hello world",
  "inputs": [
    {
      "url": "sda://DATASET-2024-012345/variants/genome2341.vcf.gz",
      "path": "/inputs/genome.vcf.gz"
    }
  ],
  "outputs": [
    {
      "url": "file:///results/line_count.txt",
      "path": "/outputs/line_count.txt"
    }
  ],
  "executors": [
    {
      "image": "alpine",
      "command": ["sh", "-c", "zcat /inputs/genome.vcf.gz | wc -l"],
      "stdout": "/outputs/line_count.txt"
    }
  ]
}
```

[sda]: https://github.com/neicnordic/sensitive-data-archive/blob/main/sda-download/api/api.md
[crypt4gh]: http://samtools.github.io/hts-specs/crypt4gh.pdf
