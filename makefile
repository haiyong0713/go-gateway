.PHONY: prow-update

prow-update:
	hack/update-prow.sh

label-update:
	label_sync -config hack/prow/labels.yaml -token /etc/gitlab/oauth -only platform/go-main -gitlab=true -confirm=true

mod-update:
	hack/update-mod.sh