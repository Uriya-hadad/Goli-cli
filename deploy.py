#!/usr/bin/python

import sys, getopt, subprocess
from os import listdir
from os import system


def main(argv):
    help_option_content = '''
deploy script help guide: python script for pushing artifacts to common artifactory
Usage: deploy.py [Global arguments...] 

Global arguments:                                   
  --help, -h                       Show help
  --password, -p                   Common Artifactoy API token 
  --bin_folder, -f                 Binaries folder name
  --bin_name, -n                   Binaries file names in the Artifactoy
  --bin_version, -v                Binaries version
  --target, -t                     Binaries remote path in Artifactoy
  
ex. python3 deploy.py -p ${COMMON_ARTIFACTORY_TOKEN} -f ${FOLDER_NAME} -n ${BINARY_NAME} -v ${VERSION} -t ${ARTIFACTORY_TARGET}
'''

    password = ''
    bin_folder = ''
    bin_name = ''
    bin_version = ''
    target = ''

    try:
        opts, args = getopt.getopt(argv, "h:p:f:n:v:t:",
                                   ["password=", "bin_folder=", "bin_name=", "bin_version=", "target="])
    except getopt.GetoptError:
        print('''Some parameters are missing, please check your script call:

Actual parameters: {}
{}'''.format(argv, help_option_content))
        sys.exit(2)
    for opt, arg in opts:
        if opt == ("-h", "--help"):
            print(help_option_content)
            sys.exit()
        elif opt in ("-p", "--password"):
            password = arg
        elif opt in ("-f", "--bin_folder"):
            bin_folder = arg
        elif opt in ("-n", "--bin_name"):
            bin_name = arg
        elif opt in ("-v", "--bin_version"):
            bin_version = arg
        elif opt in ("-t", "--target"):
            target = arg

    artifacts = listdir(bin_folder)
    print('Found artifacts for deployment: {}'.format(artifacts))

    for artifact in artifacts:

        # ex. https://common.repositories.cloud.sap/artifactory/portal/go/plugins/goli/Goli-1.0.4-linux-amd64.gz'
        #                                                             [orgPath] /   [module]  /[artifact]
        remote_artifact_path = 'https://common.repositories.cloud.sap/artifactory/{}/{}/{}'.format(target, bin_name, artifact)
        print("Trying to push: {}...".format(remote_artifact_path))
        command = ["curl",
                "-H", f"X-JFrog-Art-Api:{password}",
                "-T", f"{bin_folder}/{artifact}",
                remote_artifact_path]
        result = subprocess.run(command, capture_output=True, text=True)
        if result.returncode != 0:
            # curl failed
            print(f"failed to deploy artifact: {{artifact}} exit code: {result.returncode}")
            print(f"Output: {result.stdout}")
            print(f"Error: {result.stderr}")
            exit(1)
        else:
            print(f"Deploy {artifact} succeeded")
           

if __name__ == "__main__":
    main(sys.argv[1:])