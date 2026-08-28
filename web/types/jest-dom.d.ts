// Type augmentation for @testing-library/jest-dom matchers (toBeInTheDocument,
// toHaveTextContent, toHaveClass, ...). jest.setup.js imports the package at
// runtime, but .js setup files are outside tsconfig's typecheck scope, so the
// global matcher types never loaded for test files. This file pulls them in.
import '@testing-library/jest-dom'
