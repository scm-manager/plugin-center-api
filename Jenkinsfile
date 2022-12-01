#!/usr/bin/env groovy
def version = 'UNKNOWN'

pipeline {

  options {
    buildDiscarder(logRotator(numToKeepStr: '10'))
    disableConcurrentBuilds()
  }

  agent {
    node {
      label 'docker'
    }
  }

  stages {

    stage('Environment') {
      steps {
        script {
          def commitHashShort = sh(returnStdout: true, script: 'git rev-parse --short HEAD')
          version = "${new Date().format('yyyyMMddHHmm')}-${commitHashShort}".trim()
        }
      }
    }

    stage('Build') {
      agent {
        docker {
          image 'golang:1.17.3'
        }
      }
      environment {
        // change go cache location
        XDG_CACHE_HOME = "${WORKSPACE}/.cache"
      }
      steps {
        sh 'go build -a -tags netgo -ldflags "-w -extldflags \'-static\'" -o target/plugin-center-api *.go'
        stash name: 'target', includes: 'target/*'
      }
    }

    stage('Unit-Tests') {
      agent {
        docker {
          image 'golang:1.17.3'
        }
      }
      environment {
        // change go cache location
        XDG_CACHE_HOME = "${WORKSPACE}/.cache"
      }
      steps {
        sh 'mkdir -p target/unit-tests'
        sh 'go test -v > target/unit-tests/tests.out'
        sh 'go test -coverprofile target/unit-tests/coverage.out'
        sh 'go test -json > target/unit-tests/tests.json'
        stash name: 'testresults', includes: 'target/**'
      }
    }

    stage('Process Unit-Test Results') {
      agent {
        docker {
          image 'cloudogu/golang:1.13.10-1'
        }
      }
      environment {
        // change go cache location
        XDG_CACHE_HOME = "${WORKSPACE}/.cache"
      }
      steps {
        unstash 'testresults'
        sh 'cat target/unit-tests/tests.out | go-junit-report > target/unit-tests/unit-tests.xml'
        junit 'target/unit-tests/unit-tests.xml'
      }
    }

    stage('Sonarqube') {
      agent {
        node {
          label 'docker'
        }
      }
      environment {
        scannerHome = tool 'sonar-scanner'
      }
      steps {
        unstash 'testresults'
        withSonarQubeEnv('sonarcloud.io-scm') {
          sh "${scannerHome}/bin/sonar-scanner"
        }
        timeout(time: 10, unit: 'MINUTES') {
          waitForQualityGate abortPipeline: true
        }
      }
    }

    stage('Docker') {
      agent {
        node {
          label 'docker'
        }
      }
      steps {
        unstash 'target'
        script {
          dir("website") {
            git changelog: false, poll: false, branch: 'master', url: 'https://github.com/scm-manager/website'
          }
          docker.withRegistry('', 'cesmarvin-dockerhub-access-token') {
            def image = docker.build("scmmanager/plugin-center-api:${version}")
            image.push()
          }
        }
      }
    }

    stage('Deployment') {
      when {
        branch 'master'
      }
      agent {
        docker {
          image 'lachlanevenson/k8s-helm:v3.2.1'
          args  '--entrypoint=""'
        }
      }
      steps {
        withCredentials([file(credentialsId: 'helm-client-scm-manager', variable: 'KUBECONFIG')]) {
          sh "helm upgrade --install --set image.tag=${version} plugin-center-api helm/plugin-center-api --set oidcSecret=plugin-center-api-oidc"
        }
      }
    }

  }

  post {
    failure {
      mail to: "scm-team@cloudogu.com",
        subject: "${JOB_NAME} - Build #${BUILD_NUMBER} - ${currentBuild.currentResult}!",
        body: "Check console output at ${BUILD_URL} to view the results."
    }
  }
}
