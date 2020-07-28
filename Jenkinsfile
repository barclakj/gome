pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                script {
                    def root = tool name: 'go-1.14.6', type: 'go'

                    // Export environment variables pointing to the directory where Go was installed
                    withEnv(["GOROOT=${root}", "PATH+GO=${root}/bin"]) {
                        sh 'go version'
                    }
                }
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

