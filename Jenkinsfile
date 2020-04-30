pipeline {
    agent any
    options {
        checkoutToSubdirectory('argo-messaging')
        newContainerPerStage()
    }
    environment {
        PROJECT_DIR="argo-messaging"
        GOPATH="${WORKSPACE}/go"
        GIT_COMMIT=sh(script: "cd ${WORKSPACE}/$PROJECT_DIR && git log -1 --format=\"%H\"",returnStdout: true).trim()
        GIT_COMMIT_HASH=sh(script: "cd ${WORKSPACE}/$PROJECT_DIR && git log -1 --format=\"%H\" | cut -c1-7",returnStdout: true).trim()
        GIT_COMMIT_DATE=sh(script: "date -d \"\$(cd ${WORKSPACE}/$PROJECT_DIR && git show -s --format=%ci ${GIT_COMMIT_HASH})\" \"+%Y%m%d%H%M%S\"",returnStdout: true).trim()
    }
    stages{
        stage ('Build'){
            parallel {
                stage ('Build Centos 6') {
                    agent {
                        docker {
                            image 'argo.registry:5000/epel-6-perl'
                            args '-u jenkins:jenkins'
                        }
                    }
                    steps {
                        echo 'Building Rpm...'
                        sh """
                            cd ${WORKSPACE}/${PROJECT_DIR}
                            touch ${PROJECT_DIR}.TEST.CLEAN.tar.gz
                            """
                    }
                }
                stage ('Build Centos 7') {
                    agent {
                        docker {
                            image 'argo.registry:5000/epel-7-perl'
                            args '-u jenkins:jenkins'
                        }
                    }
                    steps {
                        echo 'Building Rpm...'
                        sh """
                            cd ${WORKSPACE}/${PROJECT_DIR}
                            touch ${PROJECT_DIR}.TEST.CLEAN.tar.gz
                            """
                    }
                }
            }
        }
    }
    post{
        always {
            cleanWs()
        }
    }
}
