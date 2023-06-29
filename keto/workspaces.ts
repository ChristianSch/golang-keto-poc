import {
    Namespace,
    Context,
} from "@ory/permission-namespace-types";

class User implements Namespace {}

class Workspace implements Namespace {
    related: {
        owners: User[];
        users: User[];
    }

    permits = {
        // all users that are owners or users of this workspace can access it
        view: (ctx: Context): boolean =>
            this.related.users.includes(ctx.subject) ||
            this.related.owners.includes(ctx.subject),
    }
}

class Unit implements Namespace {
    related: {
        // the unit is only owned by one workspace, but all relations are many-to-many in keto
        workspaces: Workspace[];
        users: User[];
    }

    permits = {
        // all workspace owners and users can view this unit
        view: (ctx: Context): boolean =>
            this.related.workspaces.traverse((w) =>
                w.related.owners.includes(ctx.subject)
            ) || this.related.users.includes(ctx.subject),
    }
}
