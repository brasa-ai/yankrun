# Transformation Functions

YankRun supports transformation functions that can be applied to placeholders to modify their values.

## Syntax

Transformations are applied by appending a colon (`:`) followed by the function name to the placeholder key. Multiple transformations can be chained together, and they will be applied in the order they are defined.

`[[PLACEHOLDER:transformation1:transformation2(arg1,arg2)]]`

## Available Functions

### `toUpperCase`

Converts the placeholder value to uppercase.

-   **Example**: `[[APP_NAME:toUpperCase]]`
-   **Input**: `my-app`
-   **Output**: `MY-APP`

### `toLowerCase` / `toDownCase`

Converts the placeholder value to lowercase. Both `toLowerCase` and `toDownCase` are supported.

-   **Example**: `[[APP_NAME:toLowerCase]]`
-   **Input**: `My-App`
-   **Output**: `my-app`

### `gsub`

Performs a global substitution on the placeholder value.

-   **Syntax**: `gsub(old,new)`
-   **`old`**: The string to be replaced.
-   **`new`**: The string to replace with.

**Examples**:

-   Replace all occurrences of `-` with `_`:
    -   **Placeholder**: `[[VERSION:gsub(-,)]]`
    -   **Input**: `1.0.0-alpha`
    -   **Output**: `1.0.0_alpha`

-   Replace all spaces with `-`:
    -   **Placeholder**: `[[PROJECT_NAME:gsub( ,-)]]`
    -   **Input**: `My Project`
    -   **Output**: `My-Project`