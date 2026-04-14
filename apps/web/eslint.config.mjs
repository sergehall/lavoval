import nextVitals from 'eslint-config-next/core-web-vitals';
import prettierConfig from 'eslint-config-prettier';

/**
 * eslint-config-prettier must be last — it disables all ESLint rules
 * that could conflict with Prettier's formatting decisions.
 */
export default [...nextVitals, prettierConfig];
