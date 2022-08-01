# Baidu Cloud Plugin

The Baidu Cloud plugin contains a single builder [baiducloud-bcc](/docs/builders/bcc.mdx) that provides for users the capability to build customized images based on an existing base images.

## Installation

### Using pre-built releases

#### Using the `packer init` command

Starting from version 1.7, Packer supports a new `packer init` command allowing
automatic installation of Packer plugins. Read the
[Packer documentation](https://www.packer.io/docs/commands/init) for more information.

To install this plugin, copy and paste this code into your Packer configuration.
Then, run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
    required_plugins {
    baiducloud = {
      version = ">= 0.0.1"
      source  = "github.com/hashicorp/baiducloud"
    }
  }
}
```

#### Manual installation

You can find pre-built binary releases of the plugin [here](https://github.com/hashicorp/packer-plugin-name/releases).
Once you have downloaded the latest archive corresponding to your target OS,
uncompress it to retrieve the plugin binary file corresponding to your platform.
To install the plugin, please follow the Packer documentation on
[installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).

#### From Source

If you prefer to build the plugin from its source code, clone the GitHub
repository locally and run the command `go build` from the root
directory. Upon successful compilation, a `packer-plugin-baiducloud` plugin
binary file can be found in the work directory.
To install the compiled plugin, please follow the official Packer documentation
on [installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).

## Plugin Contents

The Baiducloud plugin is intended as a starting point for creating Packer plugins, containing:

### Builders

- [builder](/docs/builders/bcc.mdx) - The baiducloud builder is used to create endless Packer
  plugins using a consistent plugin structure.