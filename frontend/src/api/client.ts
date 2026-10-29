import axios from 'axios'
import { ElMessage } from 'element-plus'

const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
    'X-Operator-Id': 'HR_Admin_User',
    'X-Operator-Name': 'HR 招聘专员'
  }
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    const res = response.data
    // 如果返回的是标准格式但 code !== 0，抛错
    if (res && typeof res.code === 'number' && res.code !== 0) {
      // 字段校验错误特殊处理
      if (res.code === 40001 && res.errors) {
        return Promise.reject({ isValidationError: true, errors: res.errors, message: res.message })
      }
      ElMessage.error(res.message || '请求发生错误')
      return Promise.reject(new Error(res.message || 'Error'))
    }
    return response
  },
  (error) => {
    if (error.response) {
      const status = error.response.status
      const data = error.response.data

      // Q5 乐观锁并发冲突 409
      if (status === 409) {
        ElMessage.error(data.message || '并发修改冲突：当前表单已被他人保存，版本不一致，请刷新后重试')
        return Promise.reject({ isConflict: true, status: 409, message: data.message })
      }

      if (status === 400 && data.errors) {
        return Promise.reject({ isValidationError: true, errors: data.errors, message: data.message })
      }

      ElMessage.error(data.message || `网络请求错误 [${status}]`)
    } else {
      ElMessage.error(error.message || '网络连接失败，请检查服务状态')
    }
    return Promise.reject(error)
  }
)

export default apiClient
