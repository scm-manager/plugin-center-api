#!/usr/bin/env groovy
pipeline {

  agent {
    node {
      label 'docker'
    }
  }

  stages {

    stage('Build') {
      agent {
        docker {
          image 'golang:1.12.9'
        }
      }
      environment {
        // change go cache location
        XDG_CACHE_HOME = "${WORKSPACE}/.cache"
      }
      steps {
        sh 'go build -a -tags netgo -ldflags "-w -extldflags \'-static\'" -o target/plugin-center-api *.go'
      }
    }

    stage('Unit-Tests') {
      agent {
        docker {
          image 'golang:1.12.9'
        }
      }
      environment {
        // change go cache location
        XDG_CACHE_HOME = "${WORKSPACE}/.cache"
      }
      steps {
        sh 'mkdir -p target/unit-tests'
        sh 'go test -v -coverprofile target/unit-tests/coverage.out > target/unit-tests/tests.out'
      }
    }

    stage('Process Unit-Test Results') {
      agent {
        docker {
          image 'cloudogu/golang:1.12.7-1'
        }
      }
      environment {
        // change go cache location
        XDG_CACHE_HOME = "${WORKSPACE}/.cache"
      }
      steps {
        sh 'cat target/unit-tests/tests.out | go-junit-report > target/unit-tests/unit-tests.xml'
        junit 'target/unit-tests/unit-tests.xml'
      }
    }

    stage('Sonarqube') {
      environment {
        scannerHome = tool 'sonar-scanner'
      }
      steps {
        withSonarQubeEnv('sonarcloud.io-scm') {
          sh "${scannerHome}/bin/sonar-scanner"
        }
        timeout(time: 10, unit: 'MINUTES') {
          waitForQualityGate abortPipeline: true
        }
      }
    }

    stage('Docker') {
      steps {
        script {
          def commitHashShort = sh.returnStdOut "git log -1 --pretty=%B"
          def version = "${new Date().format('yyyyMMddHHmm')}-${commitHashShort}"
          image = docker.build("scmmanager/plugin-center-api:${version}")
          docker.withRegistry('', 'hub.docker.com-cesmarvin') {
            image.push()
          }
        }
      }
    }

  }
}
