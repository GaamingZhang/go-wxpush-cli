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
                checkout scm
            }
        }

        stage('Build') {
            steps {
                echo '构建项目...'
                withEnv(["PATH+GO=/usr/local/go/bin"]) {
                    sh 'go build -o ${BINARY_NAME} main.go'
                    sh 'ls -lh ${BINARY_NAME}'
                }
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
}
