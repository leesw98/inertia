## inertia ${remote_name} user rm

Remove a user

### Synopsis

Removes the given user from Inertia's user database.

This user will no longer be able to log in and view or configure the deployment
remotely.

```
inertia ${remote_name} user rm [user] [flags]
```

### Options

```
  -h, --help   help for rm
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
  -s, --short           don't stream output from command
      --verify-ssl      verify SSL communications - requires a signed SSL certificate
```

### SEE ALSO

* [inertia ${remote_name} user](inertia_${remote_name}_user.md)	 - Configure user access to Inertia Web

