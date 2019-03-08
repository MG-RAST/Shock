pipeline {
    agent { 
        node {label 'bare-metal' }
    } 
    stages {
        stage('Build') { 
            steps {
                // Build container
                sh 'echo Build shock server'
                sh 'docker build -t mgrast/shock:testing .' 
                sh 'echo Build test client'
                sh 'docker build -t mgrast/shock-test-client -f tests/Dockerfile .'
            }
        }
        stage('Setup') {
            steps {
                // Create network
                docker network create shock-test
                // start services
                docker run --rm --network shock-test --name shock-server-mongodb --expose=27017 mongo mongod --dbpath /data/db
                docker run --rm --network shock-test --name shock-server -p 7445:7445 --link=shock-server-mongodb:mongodb mgrast/shock:testing /go/bin/shock-server
            }
        }
        stage('Test') { 
            steps {
                // execute tests
                docker run -t --network shock-test mgrast/shock-test-client  /shock-tester.sh -h http://shock-server -p 7445
                // sh 'docker run -t --rm  mgrast/shock-test-client test ' 
            }   
        }
        stage('Teardown'){
            steps{
                // shutdown container and network
                // delete images
            }
        }
    }
}
