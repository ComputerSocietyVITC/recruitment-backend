# Contributing to Recruitment Backend

## Getting Started

### 1. Fork the Repository

Before you start contributing, create your own fork of the repository:

1. Navigate to the [recruitment-backend repository](https://github.com/ComputerSocietyVITC/recruitment-backend)
2. Click the "Fork" button in the top-right corner
3. Clone your fork to your local machine:
   ```bash
   git clone https://github.com/YOUR_USERNAME/recruitment-backend.git
   cd recruitment-backend
   ```
4. Add the original repository as an upstream remote:
   ```bash
   git remote add upstream https://github.com/ComputerSocietyVITC/recruitment-backend.git
   ```

### 2. Set Up Development Environment

1. Ensure you have Go 1.19+ installed
2. Install PostgreSQL and create a database
3. Follow the setup instructions in the [README.md](README.md)
4. Verify your setup by running:
   ```bash
   make run
   ```

## Development Guidelines

### Commit Message Format

We use **semantic commit messages** to maintain a clear and consistent git history. Follow this format:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

#### Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that don't affect the meaning of the code (white-space, formatting, etc.)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools and libraries
- `ci`: Changes to CI configuration files and scripts

#### Examples:
```bash
feat(auth): add JWT token refresh functionality
fix(database): resolve connection pool timeout issue
docs(api): update authentication endpoint documentation
refactor(models): simplify user validation logic
test(utils): add comprehensive tests for JWT utilities
chore(deps): update go.mod dependencies
```

### Code Formatting and Quality

**Do not push unformatted code.** Ensure your code follows Go standards:

**Format your code** before committing:
   ```bash
   go fmt ./...
   ```

### Code Style Guidelines

- Use meaningful variable and function names
- Write clear, concise comments for public functions and complex logic
- Keep functions small and focused on a single responsibility
- Follow the existing project structure and naming conventions

## Pull Request Process

### Before Opening a PR

1. **Sync with upstream** to avoid conflicts:
   ```bash
   git fetch upstream
   git checkout master
   git merge upstream/master
   ```

2. **Create a feature branch** from master:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

3. **Make your changes** following the guidelines above

4. **Commit with semantic messages**:
   ```bash
   git add .
   git commit -m "feat(auth): add password reset functionality"
   ```

### Opening the PR

1. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a Pull Request** on GitHub with:
   - Clear title following semantic commit format
   - Brief description of changes
   - Reference to any related issues
   - Screenshots/examples if applicable

3. **PR Description Template**:
   ```markdown
   ## Description
   Brief description of changes made.

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Related Issues
   Closes #[issue_number]
   ```

## Project Structure Guidelines

When adding new features, follow the existing project structure:

- **`models/`**: Data models, DTOs, and database migrations
- **`routes/`**: API route handlers grouped by resource
- **`middleware/`**: Custom middleware functions
- **`utils/`**: Utility functions and helpers
- **`database/`**: Database connection and configuration
- **`cmds/`**: Command-line utilities

## Database Changes

For database-related changes:

1. Create appropriate migration files in `models/migrations/`
2. Follow the naming convention: `XXX_description.sql`
3. Include both up and down migrations
4. Test migrations thoroughly
5. Update documentation if schema changes affect API

## Security Considerations

- Never commit sensitive information (passwords, API keys, etc.)
- Use environment variables for configuration
- Report security vulnerabilities privately to maintainers

## Code of Conduct

Be respectful, inclusive, and professional in all interactions. We're here to build great software together!