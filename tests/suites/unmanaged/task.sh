test_unmanaged() {
	if [ "$(skip 'test_unmanaged')" ]; then
		echo "==> TEST SKIPPED: Unmanaged tests"
		return
	fi

	if [[ ${BOOTSTRAP_PROVIDER:-} == "ec2" ]]; then
		setup_awscli_credential
		# Ensure that the aws cli and juju both use the same aws region
		export AWS_DEFAULT_REGION="${BOOTSTRAP_REGION}"
	fi

	set_verbosity

	echo "==> Checking for dependencies"
	check_dependencies juju petname

	test_deploy_unmanaged
	test_spaces_unmanaged
}
