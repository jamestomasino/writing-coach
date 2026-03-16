import { Button } from '@/components/button'
import { Heading } from '@/components/heading'
import { Strong, Text, TextLink } from '@/components/text'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Register',
}

export default function Login() {
  return (
    <div className="grid w-full max-w-sm grid-cols-1 gap-8">
      <Heading>Create your account</Heading>
      <Text>
        Registration also runs through Kratos. After registration, the API maps that identity to its writing profile automatically.
      </Text>
      <Button href="/.ory/kratos/ui/registration" className="w-full">
        Open registration
      </Button>
      <Text>Already have an account? <TextLink href="/login"><Strong>Sign in</Strong></TextLink></Text>
    </div>
  )
}
