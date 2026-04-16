.PHONY: verify verify-strict verify-arch verify-config-docs verify-contract-sync verify-contract-parity verify-change-scope verify-openspec-required verify-doc-sync verify-generated verify-secrets verify-migrations verify-layout

RULES_ROOT ?= rules
RULES_ENV ?= $(RULES_ROOT)/global.env

verify:
	bash $(RULES_ROOT)/scripts/verify-config-docs.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-generated.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-secrets.sh $(RULES_ENV)

verify-strict: verify
	bash $(RULES_ROOT)/scripts/verify-contract-sync.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-contract-parity.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-openspec-required.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-change-scope.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-doc-sync.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-migrations.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-arch.sh $(RULES_ENV)
	bash $(RULES_ROOT)/scripts/verify-layout.sh $(RULES_ENV)

verify-arch:
	bash $(RULES_ROOT)/scripts/verify-arch.sh $(RULES_ENV)

verify-config-docs:
	bash $(RULES_ROOT)/scripts/verify-config-docs.sh $(RULES_ENV)

verify-contract-sync:
	bash $(RULES_ROOT)/scripts/verify-contract-sync.sh $(RULES_ENV)

verify-contract-parity:
	bash $(RULES_ROOT)/scripts/verify-contract-parity.sh $(RULES_ENV)

verify-openspec-required:
	bash $(RULES_ROOT)/scripts/verify-openspec-required.sh $(RULES_ENV)

verify-change-scope:
	bash $(RULES_ROOT)/scripts/verify-change-scope.sh $(RULES_ENV)

verify-doc-sync:
	bash $(RULES_ROOT)/scripts/verify-doc-sync.sh $(RULES_ENV)

verify-generated:
	bash $(RULES_ROOT)/scripts/verify-generated.sh $(RULES_ENV)

verify-secrets:
	bash $(RULES_ROOT)/scripts/verify-secrets.sh $(RULES_ENV)

verify-migrations:
	bash $(RULES_ROOT)/scripts/verify-migrations.sh $(RULES_ENV)

verify-layout:
	bash $(RULES_ROOT)/scripts/verify-layout.sh $(RULES_ENV)
