import { useState } from 'react'
import { authService } from '@/services/authService'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Card, CardContent } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function RegisterForm() {
  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [terms, setTerms] = useState(false)

  const validate = (): string | null => {
    if (!form.username || !form.email || !form.password || !form.confirmPassword) {
      return '请填写所有字段'
    }
    if (form.username.length < 3 || form.username.length > 50) {
      return '用户名需要3-50个字符'
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      return '请输入有效的邮箱'
    }
    if (form.password.length < 6 || form.password.length > 32) {
      return '密码需要6-32个字符'
    }
    if (form.password !== form.confirmPassword) {
      return '两次密码输入不一致'
    }
    if (!terms) {
      return '请同意用户协议和隐私政策'
    }
    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const err = validate()
    if (err) {
      setError(err)
      return
    }

    setError('')
    setLoading(true)
    try {
      await authService.register({
        username: form.username,
        email: form.email,
        password: form.password,
      })
      window.location.href = '/MainLayout'
    } catch (err: any) {
      setError(err.message || '注册失败')
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [field]: e.target.value }))
  }

  return (
    <Card>
      <CardContent className="p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="username">用户名</Label>
            <div className="relative">
              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                  <circle cx="12" cy="7" r="4"></circle>
                </svg>
              </div>
              <Input
                id="username"
                type="text"
                placeholder="请输入您的用户名"
                value={form.username}
                onChange={handleChange('username')}
                className="pl-10"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">邮箱</Label>
            <div className="relative">
              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                  <polyline points="22,6 12,13 2,6"></polyline>
                </svg>
              </div>
              <Input
                id="email"
                type="email"
                placeholder="请输入您的邮箱"
                value={form.email}
                onChange={handleChange('email')}
                className="pl-10"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">设置密码</Label>
            <div className="relative">
              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
              </div>
              <Input
                id="password"
                type="password"
                placeholder="请输入6-32位密码"
                value={form.password}
                onChange={handleChange('password')}
                className="pl-10"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirmPassword">确认密码</Label>
            <div className="relative">
              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
              </div>
              <Input
                id="confirmPassword"
                type="password"
                placeholder="请再次输入密码"
                value={form.confirmPassword}
                onChange={handleChange('confirmPassword')}
                className="pl-10"
              />
            </div>
          </div>

          <div className="flex items-start space-x-2">
            <Checkbox id="terms" checked={terms} onCheckedChange={(checked) => setTerms(checked as boolean)} />
            <Label htmlFor="terms" className="text-sm text-muted-foreground leading-tight">
              我已阅读并同意 <a href="/agreement" className="text-primary hover:underline">用户协议</a> 和 <a href="/privacy" className="text-primary hover:underline">隐私政策</a>
            </Label>
          </div>

          <Button
            type="submit"
            disabled={loading}
            className="w-full"
          >
            {loading ? '注册中...' : '立即注册'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
