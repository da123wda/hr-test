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
    if (res && typeof res.code === 'number') {
      // 成功状态码 (兼容 OkCode 1 与 0)
      if (res.code === 1 || res.code === 0) {
        return response
      }

      const msg = res.msg || res.message || '请求处理异常'
      // 字段级校验异常
      if ((res.code === 10004 || res.code === 40001) && res.errors) {
        return Promise.reject({ isValidationError: true, errors: res.errors, message: msg })
      }

      ElMessage.error(msg)
      return Promise.reject(new Error(msg))
    }
    return response
  },
  (error) => {
    if (error.response) {
      const status = error.response.status
      const data = error.response.data
      const msg = data?.msg || data?.message

      if (status === 409) {
        ElMessage.error(msg || '并发修改冲突：当前表单已被他人保存，版本不一致，请刷新后重试')
        return Promise.reject({ isConflict: true, status: 409, message: msg })
      }

      if (status === 400 && data?.errors) {
        return Promise.reject({ isValidationError: true, errors: data.errors, message: msg })
      }

      ElMessage.error(msg || `网络请求错误 [${status}]`)
    } else {
      ElMessage.error(error.message || '网络连接失败，请检查服务状态')
    }
    return Promise.reject(error)
  }
)

export default apiClient
