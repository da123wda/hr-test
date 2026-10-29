import type { FieldDefinition, FieldRule } from '@/types/schema'

export interface FieldRuntimeState {
  visible: boolean
  required: boolean
  alert: string
  disabled?: boolean
}

export class FrontendRuleEngine {
  private maxDepth: number = 20

  /**
   * 静态循环依赖检测 (DAG Cycle Detection)
   * 专门防止: A 修改 B, B 修改 C, C 又修改 A
   */
  public detectCycles(fields: FieldDefinition[]): { hasCycle: boolean; path: string[] } {
    const adj = new Map<string, string[]>()
    const allNodes = new Set<string>()

    // 提取依赖关系
    for (const f of fields) {
      allNodes.add(f.key)
      if (!f.rules || f.rules.length === 0) continue

      for (const rule of f.rules) {
        allNodes.add(rule.targetKey)
        const sources = this.extractSourceKeys(rule.expression, fields)
        const effectiveSources = sources.length > 0 ? sources : [f.key]
        for (const src of effectiveSources) {
          if (!adj.has(src)) {
            adj.set(src, [])
          }
          adj.get(src)!.push(rule.targetKey)
        }
      }
    }

    // 0: unvisited, 1: visiting, 2: visited
    const color = new Map<string, number>()
    const path: string[] = []
    let cyclePath: string[] = []

    const dfs = (u: string): boolean => {
      color.set(u, 1)
      path.push(u)

      const neighbors = adj.get(u) || []
      for (const v of neighbors) {
        if (color.get(v) === 1) {
          const idx = path.indexOf(v)
          cyclePath = path.slice(idx)
          cyclePath.push(v)
          return true
        }
        if (!color.get(v)) {
          if (dfs(v)) return true
        }
      }

      color.set(u, 2)
      path.pop()
      return false
    }

    for (const node of allNodes) {
      if (!color.get(node)) {
        if (dfs(node)) {
          return { hasCycle: true, path: cyclePath }
        }
      }
    }

    return { hasCycle: false, path: [] }
  }

  /**
   * 动态求值并返回所有字段的渲染状态
   */
  public evaluateRules(
    fields: FieldDefinition[],
    formState: Record<string, any>
  ): { states: Record<string, FieldRuntimeState>; cycleError?: string } {
    // 1. 环检测
    const cycleCheck = this.detectCycles(fields)
    if (cycleCheck.hasCycle) {
      const errorMsg = `[RuleEngine 循环依赖告警] 规则存在闭环依赖: ${cycleCheck.path.join(' -> ')}，已自动拦截以保护界面！`
      console.warn(errorMsg)
      return {
        states: this.initDefaultStates(fields),
        cycleError: errorMsg
      }
    }

    // 2. 初始化各字段默认状态
    const states = this.initDefaultStates(fields)

    // 3. 运行联动规则
    let stepCount = 0
    for (const f of fields) {
      if (!f.rules || f.rules.length === 0) continue

      for (const rule of f.rules) {
        stepCount++
        if (stepCount > this.maxDepth) {
          console.warn(`[RuleEngine] 联动步数超过最大深度限制 ${this.maxDepth}，已安全截断`)
          break
        }

        const match = this.evaluateExpression(rule.expression, formState)
        const targetState = states[rule.targetKey]
        if (!targetState) continue

        if (match) {
          switch (rule.action) {
            case 'hide':
              targetState.visible = false
              break
            case 'show':
              targetState.visible = true
              break
            case 'setRequired':
              targetState.required = true
              break
            case 'unsetRequired':
              targetState.required = false
              break
            case 'showAlert':
              targetState.alert = rule.parameters?.text || '触发联动审批提示'
              break
          }
        }
      }
    }

    return { states }
  }

  private initDefaultStates(fields: FieldDefinition[]): Record<string, FieldRuntimeState> {
    const states: Record<string, FieldRuntimeState> = {}
    for (const f of fields) {
      states[f.key] = {
        visible: true,
        required: !!f.validation?.required,
        alert: ''
      }
    }
    return states
  }

  /**
   * 简易且安全的条件表达式求值器
   */
  public evaluateExpression(expr: string, data: Record<string, any>): boolean {
    if (!expr) return true
    const trimmed = expr.trim()
    if (trimmed === 'true') return true
    if (trimmed === 'false') return false

    // 支持 &&
    if (trimmed.includes('&&')) {
      const parts = trimmed.split('&&')
      return parts.every((p) => this.evaluateExpression(p, data))
    }

    // 支持 ||
    if (trimmed.includes('||')) {
      const parts = trimmed.split('||')
      return parts.some((p) => this.evaluateExpression(p, data))
    }

    // 匹配操作符
    const ops = ['>=', '<=', '==', '!=', '>', '<']
    let foundOp = ''
    for (const op of ops) {
      if (trimmed.includes(op)) {
        foundOp = op
        break
      }
    }

    if (!foundOp) {
      const v = data[trimmed]
      return v !== undefined && v !== null && v !== false && v !== '' && v !== 0
    }

    const [leftKey, rightRaw] = trimmed.split(foundOp).map((s) => s.trim())
    const leftVal = data[leftKey]
    const rightClean = rightRaw.replace(/^['"]|['"]$/g, '')

    // 数值比较
    const leftNum = Number(leftVal)
    const rightNum = Number(rightClean)
    if (!isNaN(leftNum) && !isNaN(rightNum) && leftVal !== '' && leftVal !== null) {
      switch (foundOp) {
        case '==':
          return leftNum === rightNum
        case '!=':
          return leftNum !== rightNum
        case '>':
          return leftNum > rightNum
        case '<':
          return leftNum < rightNum
        case '>=':
          return leftNum >= rightNum
        case '<=':
          return leftNum <= rightNum
      }
    }

    // 字符串比较
    const leftStr = String(leftVal ?? '')
    switch (foundOp) {
      case '==':
        return leftStr === rightClean
      case '!=':
        return leftStr !== rightClean
      case '>':
        return leftStr > rightClean
      case '<':
        return leftStr < rightClean
      case '>=':
        return leftStr >= rightClean
      case '<=':
        return leftStr <= rightClean
    }

    return false
  }

  private extractSourceKeys(expr: string, fields: FieldDefinition[]): string[] {
    const identRegex = /[a-zA-Z_][a-zA-Z0-9_]*/g
    const matches = expr.match(identRegex) || []
    const fieldMap = new Set(fields.map((f) => f.key))
    const res = new Set<string>()

    for (const m of matches) {
      if (fieldMap.has(m)) {
        res.add(m)
      }
    }
    return Array.from(res)
  }
}
