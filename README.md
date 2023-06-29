# Zanzibar-inspired Permission POC

Open Ends:

- [ ] implement the workspace-> unit relation

## Usage

Run the docker containers:

```bash
cd keto && docker-compose up -d
```

And run the API:

```bash
go run cmd/server/main.go
```

Until [this issue](https://github.com/ory/keto/issues/1339) is fixed, you need to run all `curl http://0.0.0.0:4467/admin/relation-tuples` commands listed below to create the relations.

## Permissions
We can define a basic [permission model](https://www.ory.sh/docs/keto/modeling/create-permission-model)
as follows. We have the following entities for now:

- workspaces
- users
- units

Users can be owners or users of workspaces, and a workspace contains units (exclusively).

We can define the following permissions in OPL as such:

Users have access to specific workspaces. We'll use the Zanzibar relation notation as defined in the paper:

```
⟨tuple⟩ ::= ⟨object⟩‘#’⟨relation⟩‘@’⟨user⟩
⟨object⟩ ::= ⟨namespace⟩‘:’⟨object id⟩
⟨user⟩ ::= ⟨user id⟩ | ⟨userset⟩
⟨userset⟩ ::= ⟨object⟩‘#’⟨relation⟩
```

Which gives us:

- `Workspace:1#owners:@alice`
- `Workspace:1#users:@bob`
- `Workspace:2#users:@alice`
- `Workspace:1#view@relation:owners#member`: all owners also are permitted to view the workspace
- `Workspace:1#view@relation:users#member`: all users also are permitted to view the workspace

More explicitly:

- `Workspace:1#view:@alice`
- `Workspace:1#view:@bob`
- `Workspace:2#view:@alice`

This essentially means that alice is an owner of workspace 1, and a user of workspace 2. bob is a user of workspace 1. Both alice and bob can view workspace 1, but only alice can view workspace 2.

We can formulate this in typescript/OLP as such:
    
```typescript
import { Namespace, SubjectSet, Context } from "@ory/permission-namespace-types"

class User implements Namespace {}

class Workspace implements Namespace {
    related: {
        owners: User[]
        users: User[]
    }

    permits = {
        // all users that are owners or users of this workspace can access it
        view: (ctx: Context): boolean => {
            return this.related.users.includes(ctx.subject) ||
            this.related.owners.includes(ctx.subject)
        },
    }
}

class Unit implements Namespace {
    related: {
        // the unit is only owned by one workspace, but all relations are many-to-many in keto
        workspaces: Workspace[]
        users: User[]
    }

    permit = {
        // all workspace owners and users can view this unit
        view: (ctx: Context): boolean => {
            return this.related.workspace.related.owners.includes(ctx.subject) ||
            this.related.users.includes(ctx.subject)
        }
    }
}
```

This will be loaded via the keto.yml config.

## Relations
To define relations which we can check later, we'll talk to keto directly. We have alice as the owner of `workspace:1` and user of `workspace:2`, and bob as user of `workspace:1`.

```bash
~/d/k/keto >>>  curl http://0.0.0.0:4467/admin/relation-tuples -d ' { 
                    "$schema": "../.schema/relation_tuple.schema.json",
                    "namespace": "Workspace",
                    "object": "1",
                    "relation": "owners",
                    "subject_set": {
                        "namespace": "User",
                        "object": "alice"
                    }
                }' -X PUT
{"namespace":"Workspace","object":"1","relation":"owners","subject_set":{"namespace":"User","object":"alice","relation":""}}
```

```bash
~/d/k/keto >>>  curl http://0.0.0.0:4467/admin/relation-tuples -d ' {
                    "$schema": "../.schema/relation_tuple.schema.json",
                    "namespace": "Workspace",
                    "object": "2",
                    "relation": "users",
                    "subject_set": {
                        "namespace": "User",
                        "object": "alice"
                    }
                }' -X PUT
```

```bash
~/d/k/keto >>>  curl http://0.0.0.0:4467/admin/relation-tuples -d '{
                    "$schema": "../.schema/relation_tuple.schema.json",
                    "namespace": "Workspace",
                    "object": "1",
                    "relation": "users",
                    "subject_set": {
                        "namespace": "User",
                        "object": "bob"
                    }
                }' -X PUT
```

Verify this works. alice is an owner:
```bash
~/d/k/keto >>>  curl "http://0.0.0.0:4466/relation-tuples/check" -d '{
                "namespace": "Workspace",
                "object":"1",
                "relation":"owners",
                "subject_set":{"namespace":"User","object":"alice"}
                }' -X POST
{"allowed":true}
```

bob is not an owner:
```bash
~/d/k/keto >>>  curl "http://0.0.0.0:4466/relation-tuples/check" -d '{
                "namespace": "Workspace",
                "object":"1",
                "relation":"owners",
                "subject_set":{"namespace":"User","object":"bob"}
                }' -X POST
{"allowed":false}
```

but a user:
```bash
~/d/k/keto >>>  curl "http://0.0.0.0:4466/relation-tuples/check" -d '{
                "namespace": "Workspace",
                "object":"1",
                "relation":"users",
                "subject_set":{"namespace":"User","object":"bob"}
                }' -X POST
{"allowed":true}
```

How do we find out what workspaces alice has access to?

```bash
~/d/k/keto >>>  curl "http://0.0.0.0:4466/relation-tuples?namespace=Workspace&subject_set.namespace=User&subject_set.object=alice&subject_set.relation" | jq
{
  "relation_tuples": [
    {
      "namespace": "Workspace",
      "object": "1",
      "relation": "owners",
      "subject_set": {
        "namespace": "User",
        "object": "alice",
        "relation": ""
      }
    },
    {
      "namespace": "Workspace",
      "object": "2",
      "relation": "users",
      "subject_set": {
        "namespace": "User",
        "object": "alice",
        "relation": ""
      }
    }
  ],
  "next_page_token": ""
}
```