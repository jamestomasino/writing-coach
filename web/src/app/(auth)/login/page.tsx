import { Button } from '@/components/button'
import { Heading } from '@/components/heading'
import { Strong, Text, TextLink } from '@/components/text'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Login',
}

export default function Login() {
  return (
    <div className="grid w-full max-w-sm grid-cols-1 gap-8">
      <Heading>Sign in to your account</Heading>
      <Text>
        Authentication runs through Ory Kratos. Use the browser flow below so the app and API share the same session cookies.
      </Text>
      <Button href="/.ory/kratos/ui/login" className="w-full">
        Open sign in
      </Button>
      <Text>Don’t have an account? <TextLink href="/register"><Strong>Register</Strong></TextLink></Text>
    </div>
  )
}
