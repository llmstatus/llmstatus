# Ansible deployment

All deployments to non-dev environments are driven from this directory.
Direct SSH / ad-hoc `docker compose` on production servers is forbidden
(see `CONTRIBUTING.md` and the private CLAUDE.md §12).

## Layout

```
deploy/ansible/
├── inventories/
│   ├── dev/hosts.yml        # local dev (mostly unused; dev runs in Docker Compose)
│   ├── staging/hosts.yml
│   └── prod/hosts.yml
├── roles/                   # reusable role definitions (LLMS-003+)
└── playbooks/               # top-level playbooks: deploy.yml, rollback.yml, etc.
```

## Running

```bash
cd deploy/ansible
ansible-playbook -i inventories/staging/hosts.yml playbooks/deploy.yml --check
ansible-playbook -i inventories/staging/hosts.yml playbooks/deploy.yml
```

## Secrets

Secrets (API keys, DB passwords) live in Ansible Vault. The vault
password is stored in the operator's local `~/.ansible-vault-password`
file (0600). Never commit the unencrypted vault or the password.

## When something breaks

If a playbook fails, fix the playbook / role / template and re-run.
**Do not SSH in to patch the server by hand.** File a GitHub issue under
the `deploy` label if the root cause is not obvious.

The single narrow exception is a migration "dirty state" recovery
(`migrate force N`), which may be performed manually on the database.
Everything else stays in Ansible.
