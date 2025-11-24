pipeline {
    agent {
        docker {
            image 'argo.registry:5000/rocky9-go1.25:latest'
            args '-u jenkins:jenkins'
        }
    }
    options {
        checkoutToSubdirectory('argo-messaging')
        newContainerPerStage()
    }
    environment {
        PROJECT_DIR="argo-messaging"
        GH_USER = 'newgrnetci'
        GH_EMAIL = '<argo@grnet.gr>'
        GOCACHE = '/tmp/go-cache'
        GOMODCACHE = '/tmp/go-mod-cache'
        GOPATH="${WORKSPACE}/go"
        GIT_COMMIT=sh(script: "cd ${WORKSPACE}/$PROJECT_DIR && git log -1 --format=\"%H\"",returnStdout: true).trim()
        GIT_COMMIT_HASH=sh(script: "cd ${WORKSPACE}/$PROJECT_DIR && git log -1 --format=\"%H\" | cut -c1-7",returnStdout: true).trim()
        GIT_COMMIT_DATE=sh(script: "date -d \"\$(cd ${WORKSPACE}/$PROJECT_DIR && git show -s --format=%ci ${GIT_COMMIT_HASH})\" \"+%Y%m%d%H%M%S\"",returnStdout: true).trim()
    }
    stages {
        stage('Build') {
            steps {
                echo 'Build...'
                sh """
                go version
                mkdir -p ${WORKSPACE}/go/src/github.com/ARGOeu
                ln -sf ${WORKSPACE}/${PROJECT_DIR} ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                rm -rf ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}/${PROJECT_DIR}
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                export CGO_CFLAGS"=-O2 -fstack-protector --param=ssp-buffer-size=4 -D_FORTIFY_SOURCE=2"
                go build -buildmode=pie -ldflags "-s -w -linkmode=external -extldflags '-z relro -z now'"
                """
            }
        }
        stage('Security Tests') {
            steps {
                sh """
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}

                checksec --file=./argo-messaging --format=xml > ./checksec.xml

                set +x
                # define function that receives field/value and checks them in checksec.xml output
                checksec_point(){ f=\$1; v=\$2; r=\$(xmllint --xpath "string(//file/@\$f)" checksec.xml); \
                echo -n "\$f(expected:\$v)=\$r"; [[ "\$r" == "\$v" ]] && \
                echo -e "\t✓ PASS" || { echo -e "\t𐄂 FAIL"; return 1; }; }

                # for pairs of field/value items check if they exist in the checksec.xml output - break if not
                for pair in "pie yes" "nx yes" "relro full" "rpath no" "runpath no" "symbols no" "fortify_source yes"; \
                do set -- \$pair; checksec_point "\$1" "\$2"; done
                """
            }
        }
        stage('Test') {
            steps {
                echo 'Test & Coverage...'
                sh """
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                gotestsum --junitfile ${WORKSPACE}/junit.xml -- -p 1 -v -coverprofile=coverage.out ./...
                gocover-cobertura < coverage.out > ${WORKSPACE}/coverage.xml
                """
                junit '**/junit.xml'
                publishCoverage adapters: [
                    coberturaAdapter('**/coverage.xml')
                ]
            }
        }
        stage('Package') {
            steps {
                echo 'Building Rpm...'
                withCredentials(bindings: [sshUserPrivateKey(credentialsId: 'jenkins-rpm-repo', usernameVariable: 'REPOUSER', \
                                                             keyFileVariable: 'REPOKEY')]) {
                    sh "/home/jenkins/build-rpm.sh -w ${WORKSPACE} -b ${BRANCH_NAME} -d rocky9 -p ${PROJECT_DIR} -s ${REPOKEY}"
                }
                archiveArtifacts artifacts: '**/*.rpm', fingerprint: true
            }
            post{
                always {
                    cleanWs()
                }
            }
        }
    }
    post{
        always {
            cleanWs()
        }
        success {
            script{
                if ( env.BRANCH_NAME == 'devel' ) {
                    build job: '/ARGO/argodoc/devel', propagate: false
                } else if ( env.BRANCH_NAME == 'master' ) {
                    build job: '/ARGO/argodoc/master', propagate: false
                }
                if ( env.BRANCH_NAME == 'master' || env.BRANCH_NAME == 'devel' ) {
                    slackSend( message: ":rocket: New version for <$BUILD_URL|$PROJECT_DIR>:$BRANCH_NAME Job: $JOB_NAME !")
                }
            }
        }
        failure {
            script{
                if ( env.BRANCH_NAME == 'master' || env.BRANCH_NAME == 'devel' ) {
                    slackSend( message: ":rain_cloud: Build Failed for <$BUILD_URL|$PROJECT_DIR>:$BRANCH_NAME Job: $JOB_NAME")
                }
            }
        }
    }
}
