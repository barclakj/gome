pipeline {
agent {
// Run on an agent where we want to use Go
node {
    // Install the desired Go version
    def root = tool name: 'go-1.14.6', type: 'go'

    // Export environment variables pointing to the directory where Go was installed
    withEnv(["GOROOT=${root}", "PATH+GO=${root}/bin"]) {
        sh 'go version'
    }
}
}

    stages {
        stage('Build') {
            steps {
                echo 'Building..'
		sh 'go build'
            }
        }
        stage('Test') {
            steps {
                echo 'Testing..'
            }
        }
        stage('Deploy') {
            steps {
                echo 'Deploying....'
            }
        }
    }
}

