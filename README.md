# smurfd

Simple AWS Secrets Manager CLI Client

Returns AWS Secrets Manager secrets in simple formatting based on the desire from the user.

Useful for when you want to grab secrets and you don't want to deal with the full AWS CLI and all its nasty python dependencies or if you just want a lightweight binary to distribute to docker images, etc. 

## Usage
Here's the general usage help from the CLI.
```
Usage of ./smurfd:
  -config string
    	Filename of YAML configuration file. Contents overrides all parameters. Leave blank to use parameters only.
  -profile string
    	AWS session credentials profile, if blank default or instance profile will be attempted.
  -raw
    	pull down the raw value of the secret instead of tag parsing.
  -secretname string
    	name of secret to retrieve (default "foo")
  -tag string
    	the tag key name to grab from the secret. The value of this key will be returned (default "username")
```

Say you have a secret stored named `my-bank-account` and the value is list of key/values like so:

```json
[
    {
    "Key": "username",
    "Value": "hackerman"
    },
    {
    "Key": "password",
    "Value": "hunter1"
    },
    {
    "Key": "accountnumber",
    "Value": "12345"
    }
]
```

You would retrieve the password and the username in two separate commands like this:
```
$ ./smurfd -secretname my-bank-account -tag password
hunter1
$ ./smurfd -secretname my-bank-account -tag username
hackerman
```

You can also pass in a profile entry name if you're not using default or instance profile credentials.
```
$ ./smurfd -profile contoso-prod -secretname my-bank-account -tag password
hunter1
```
## Secret Requirements (tag mode)
All secrets have to be stored in the `[{'Key':'foo', 'Value':'bar'}]` format in order for `smurfd` to parse it properly.

## Raw Mode
Alternatively you can just grab the entire secret and dump it to standard out by passing the `-raw` parameter like so:

```
$ ./smurfd -profile default -secretname iamthelaw-config -raw | jq .
{
  "Project": {
    "AppName": "iam-the-law",
    "MajorVersion": 0,
    "MinorVersion": 0,
    "SlackChannel": "#cloudpod-feed-dev"
  }
}
```

## IAM Requirements
The following policy is the bare minimum for being able to retrieve the secret and decrypt it. Obviously the ARN for your KMS key needs to be updated.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowSecretsManager",
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetSecretValue",
                "secretsmanager:DescribeSecret"
            ],
            "Resource": "*"
        },
        {
            "Sid": "AllowKMSDecryptMaster",
            "Effect": "Allow",
            "Action": "kms:Decrypt",
            "Resource": "arn:aws:kms:us-east-1:123456789012:key/keyid"
        }
    ]
}
```

## Jenkins Pipeline Usage
The following groovy pipeline script pulls down the secret named `terradacter-deploy-key` and stores it in the `${PPASS}` variable for use later on in the pipeline. 

Obviously, this assumes that the `smurfd` binary has been deployed to an accessible path on the Jenkins server (e.g., `/usr/bin/smurfd`).

For available binaries see the releases tab in this repository.

```
try {
    node("master"){
        stage ("grabbing") {
            PPASS = sh (
                script: 'smurfd -secretname terradacter-deploy-key -tag password',
                returnStdout: true
            ).trim()
            echo "Dat pass: ${PPASS}"
        }
    }
} catch (error) {
    print error
}
```

The above pipeline has the following output:

```
Started by user Russell Endicott
Running in Durability level: MAX_SURVIVABILITY
[Pipeline] node
Running on Jenkins in /var/lib/jenkins/workspace/gaudi/grabber
[Pipeline] {
[Pipeline] stage
[Pipeline] { (grabbing)
[Pipeline] sh
[grabber] Running shell script
+ smurfd -secretname terradacter-deploy-key -tag password
[Pipeline] echo
Dat pass: 5c46nevergonnagiveyouupabf5381231501206
[Pipeline] }
[Pipeline] // stage
[Pipeline] }
[Pipeline] // node
[Pipeline] End of Pipeline
Finished: SUCCESS
```

## Config File
If you don't want to pass in parameters via CLI you can store them in a YAML config file and pass that in as the only parameter. Its contents will override any command line options. See `sample-config.yml` for documentation.
