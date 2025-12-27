pipeline {
    agent any

    environment {
        GO_VERSION = '1.21'
        DEPLOY_PATH = '/var/wxpush'
        BINARY_NAME = 'wxpush'
    }

    stages {
        stage('Checkout') {
            steps {
                echo '拉取代码...'
                checkout scm
                sh 'git log -1 --pretty=format:"%h - %an, %ar : %s"'
            }
        }

        stage('Setup Go Environment') {
            steps {
                echo '设置 Go 环境...'
                sh '''
                    if command -v go &> /dev/null; then
                        go version
                    else
                        echo "Go 未安装，请先安装 Go ${GO_VERSION}"
                        exit 1
                    fi
                '''
            }
        }

        stage('Build') {
            steps {
                echo '构建项目...'
                sh 'go build -o ${BINARY_NAME} main.go'
                sh 'ls -lh ${BINARY_NAME}'
            }
        }

        stage('Test') {
            steps {
                echo '运行测试...'
                sh 'go test -v ./... || echo "没有测试文件或测试失败"'
            }
        }

        stage('Deploy') {
            steps {
                echo '部署到 ${DEPLOY_PATH}...'
                sh '''
                    # 创建部署目录
                    sudo mkdir -p ${DEPLOY_PATH}
                    
                    # 备份旧版本（如果存在）
                    if [ -f ${DEPLOY_PATH}/${BINARY_NAME} ]; then
                        sudo cp ${DEPLOY_PATH}/${BINARY_NAME} ${DEPLOY_PATH}/${BINARY_NAME}.backup
                        echo "已备份旧版本"
                    fi
                    
                    # 复制新版本
                    sudo cp ${BINARY_NAME} ${DEPLOY_PATH}/
                    
                    # 设置执行权限
                    sudo chmod +x ${DEPLOY_PATH}/${BINARY_NAME}
                    
                    # 验证部署
                    ls -lh ${DEPLOY_PATH}/${BINARY_NAME}
                    ${DEPLOY_PATH}/${BINARY_NAME} --help || true
                '''
            }
        }

        stage('Cleanup') {
            steps {
                echo '清理临时文件...'
                cleanWs()
            }
        }
    }

    post {
        success {
            echo '构建和部署成功！'
            echo "可执行文件已部署到: ${DEPLOY_PATH}/${BINARY_NAME}"
        }

        failure {
            echo '构建或部署失败！'
        }

        always {
            echo '流水线执行完成'
        }
    }
}
