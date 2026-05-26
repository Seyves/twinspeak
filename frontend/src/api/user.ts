import { httpClient } from '@/lib/httpClient'

export type UserInfo = {
    email: string
    profilePicture: string | null
    createdAt: string
}

export async function getMe(): Promise<UserInfo> {
    const response = await httpClient.get('me')
    return response.json()
}
