# TF
Terraform configuration for the Plus ID utility.

## Requirements

### 0. Create your private Notion Extension
You can create your Internal Notion Extension [here](https://www.notion.so/my-integrations).

The only two required `Content Capabilities` are:
- Read Content
- Update Content

Everything else can be optionally disabled.

Make sure you store the `Internal Integration Token` for later.

### 1. Install Terraform
[How to install Terraform](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/install-cli)

Just make sure your Terraform version is >1.3.5.
### 2. Create Google Cloud Platform project
- Initialise an empty and free GCP project at [cloud.google.com](https://cloud.google.com).
- [Install GCloud CLI](https://cloud.google.com/sdk/docs/install).
- Make sure your GCloud CLI is initialised and you're logged in with the access to your GCP project.

### 3. Create an empty GCP Cloud Storage bucket
This bucket will be used to store the Terraform state.

```bash
gcloud storage buckets create gs://$BUCKET_NAME
```
where `$BUCKET_NAME` is a name of the bucket you would like to use to store your terraform state.

### 4. Set configuration
You need to provide the configuration for the first time run.

Use the example configurations to make sure that terraform config can be properly applied.

```bash
cp backend.conf.example backend.conf
cp config.yaml.example config.yaml
```

Within your `backend.conf` make sure to change `bucket` value to the name of created bucket in step 3.
In `config.yaml`, change the `project` value to your GCP project ID.

### 5. Apply
Simply run `run.sh`.

```bash
bash run.sh
```

The script would prompt you for `Notion Client Secret`. Just provide the value of `Internal Integration Token` from step 0.

If you would like to reapply/change the terraform configuration, simply rerun the `run.sh`.

After everything is applied, you can now move to the installation of the App Engine component.
