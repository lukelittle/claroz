from django import forms
from django.contrib.auth.forms import UserCreationForm
from django.contrib.auth.models import User
from .models import Profile

class SignUpWithFederationForm(UserCreationForm):
    federation_server = forms.URLField(
        required=False,
        widget=forms.URLInput(attrs={'class': 'form-control'}),
        help_text="Optional: URL of your AT Protocol identity server"
    )
    federation_handle = forms.CharField(
        max_length=255,
        required=False,
        widget=forms.TextInput(attrs={'class': 'form-control'}),
        help_text="Optional: Your handle on the federated server"
    )
    federation_did = forms.CharField(
        max_length=255,
        required=False,
        widget=forms.TextInput(attrs={'class': 'form-control'}),
        help_text="Optional: Your DID on the federated server"
    )
    email = forms.EmailField(
        max_length=254,
        required=True,
        widget=forms.EmailInput(attrs={'class': 'form-control'})
    )

    class Meta:
        model = User
        fields = ('username', 'email', 'password1', 'password2', 
                 'federation_server', 'federation_handle', 'federation_did')
        
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        # Add Bootstrap classes to the default fields
        self.fields['username'].widget.attrs.update({'class': 'form-control'})
        self.fields['password1'].widget.attrs.update({'class': 'form-control'})
        self.fields['password2'].widget.attrs.update({'class': 'form-control'})

    def save(self, commit=True):
        user = super().save(commit=True)  # Save the user first to create profile
        
        # Update profile with federation details
        profile = user.profile
        profile.federation_server = self.cleaned_data.get('federation_server')
        profile.federation_handle = self.cleaned_data.get('federation_handle')
        profile.federation_did = self.cleaned_data.get('federation_did')
        
        if commit:
            profile.save()
            
        return user
