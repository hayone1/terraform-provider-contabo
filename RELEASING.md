# How to release the provider

This guide explains how to publish a new version of the `hayone1/contabo` provider. The first part is a one-time setup. The second part repeats for every release.

## Part 1: One-time setup

### Step 1: Create a GPG signing key

The Terraform Registry and the OpenTofu Registry only accept a release when a GPG key signs the release. The Terraform Registry accepts RSA keys, so create an RSA key with 4096 bits.

Git for Windows includes GPG, so you can run these commands in Git Bash. On Linux or in WSL, GPG is usually installed already.

1. Start the key generator:

   ```sh
   gpg --full-generate-key
   ```

2. Answer the questions of the key generator like this:
   - Choose the key type `RSA and RSA`.
   - Choose the key size `4096`.
   - Choose how long the key is valid. A key that never expires is the simplest choice. A key that expires is safer, but you must extend the key before the key expires.
   - Enter your name and your email address.
   - Enter a passphrase. Keep the passphrase, because the release workflow needs the passphrase.

3. Find the id of the new key. The id is the long hexadecimal value in the `sec` line of this command:

   ```sh
   gpg --list-secret-keys --keyid-format=long
   ```

   The rest of this guide calls this value `KEY_ID`.

### Step 2: Export the public key and the private key

1. Export the public key. The registries need the public key to check the signatures:

   ```sh
   gpg --armor --export KEY_ID > contabo-provider-public.asc
   ```

2. Export the private key. The release workflow needs the private key to sign the releases:

   ```sh
   gpg --armor --export-secret-keys KEY_ID > contabo-provider-private.asc
   ```

Keep the private key file secret. You delete the private key file in step 3.

### Step 3: Add the secrets to the GitHub repository

The release workflow reads two secrets: `GPG_PRIVATE_KEY` and `PASSPHRASE`. Add both secrets to the repository `hayone1/terraform-provider-contabo`:

```sh
gh secret set GPG_PRIVATE_KEY -R hayone1/terraform-provider-contabo < contabo-provider-private.asc
gh secret set PASSPHRASE -R hayone1/terraform-provider-contabo
```

The second command asks for the passphrase. Type the passphrase from step 1.

You can also add the secrets on the GitHub website. Open the repository, then open **Settings**, **Secrets and variables**, **Actions**, and click **New repository secret**.

When both secrets exist, delete the private key file:

```sh
rm contabo-provider-private.asc
```

The private key also stays in your GPG keyring, so you can export the private key again later.

### Step 4: Add the public key to the Terraform Registry

1. Open [registry.terraform.io](https://registry.terraform.io) and click **Sign in**. Sign in with your GitHub account `hayone1`.
2. Open **User Settings** and then **Signing Keys**.
3. Click **New GPG Key**. Choose the namespace `hayone1`, and paste the whole content of `contabo-provider-public.asc`, including the `BEGIN` line and the `END` line.
4. Click **Save**.

Add the key before the first release. The registry checks every release against the keys of the namespace.

## Part 2: Every release

### Step 5: Prepare the release

1. Make sure that the `main` branch is up to date and that the CI workflow on GitHub passed for the last commit.
2. In `CHANGELOG.md`, replace "(not released yet)" in the heading of the new version with the release date, for example `## v0.2.0 (2026-11-10)`.
3. Commit the change and push the change to `main`.

### Step 6: Create the tag

The release workflow starts when a tag that begins with `v` reaches GitHub:

```sh
git tag -a v0.2.0 -m "v0.2.0"
git push origin v0.2.0
```

Do not use tags like `v0.2.0-rc1` for normal releases. Terraform treats such a tag as a pre-release and does not install a pre-release unless the configuration asks for that exact version.

### Step 7: Check the release workflow

1. Open the **Actions** tab of the repository and open the run of the **release** workflow.
2. Wait until the run finishes. The run tests the provider, builds the binaries, signs the checksums and creates a GitHub release.
3. Open the new release on the **Releases** page. The release must contain the zip files, a `SHA256SUMS` file, a `SHA256SUMS.sig` file and a `manifest.json` file.

If the run fails in the **Import GPG key** step, one of the two secrets is wrong. Fix the secret, delete the tag with `git push --delete origin v0.2.0` and `git tag -d v0.2.0`, and create the tag again.

### Step 8: Publish the provider on the Terraform Registry (only for the first release)

1. Open [registry.terraform.io](https://registry.terraform.io) and sign in.
2. Click **Publish** and then **Provider**.
3. Choose the repository `hayone1/terraform-provider-contabo` and accept the terms.
4. Click **Publish Provider**.

The registry reads the existing release and creates the provider page from the `docs/` folder. For later releases, you do not need this step. GitHub informs the registry about every new release, and the registry adds the new version within a few minutes.

### Step 9: Publish the provider on the OpenTofu Registry (only for the first release)

The OpenTofu Registry adds providers through GitHub issues. Use the web form on GitHub, because the registry does not process issues that someone creates through the API.

1. Open [github.com/opentofu/registry/issues/new/choose](https://github.com/opentofu/registry/issues/new/choose).
2. Choose **Submit new Provider**. Enter the repository `hayone1/terraform-provider-contabo` and submit the issue.
3. Go back to the list of issue templates and choose **Submit new Provider Signing Key**. Enter the namespace `hayone1`, paste the content of `contabo-provider-public.asc`, and submit the issue.

The OpenTofu team or their automation processes the issues. The registry then picks up new releases automatically.

### Step 10: Check the published provider

Create an empty folder with a file `main.tf` and this content:

```terraform
terraform {
  required_providers {
    contabo = {
      source  = "hayone1/contabo"
      version = "0.2.0"
    }
  }
}
```

Run `terraform init` in the folder. Terraform must download the provider and show that the signature is valid. When the OpenTofu Registry has added the provider, run `tofu init` in the same folder to check the OpenTofu Registry too.
