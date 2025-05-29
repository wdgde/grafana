package v0alpha1

UserSpec: {
    name: string
    login: string
    email: string
    emailVerified: bool
    disabled: bool
    internalID: int64 @json("-")
}