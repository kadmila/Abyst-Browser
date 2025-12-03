# Generate timestamp in the format YYYYMMddHHmm
$timestamp = (Get-Date).ToString("yyyyMMddHHmm")

# Create the tag name
$tag = "v0.1.$timestamp"

# Print out the tag
Write-Host "Creating git tag: $tag"

# Run git tag
git tag $tag