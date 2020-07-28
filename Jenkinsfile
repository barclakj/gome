pipeline {
    agent any

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

