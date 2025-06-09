import { AST_NODE_TYPES, TSESTree } from '@typescript-eslint/utils';
import { RuleContext, RuleFixer } from '@typescript-eslint/utils/dist/ts-eslint';

type TranslationRuleContext = RuleContext<
  'noUntranslatedStrings' | 'noUntranslatedStringsProp' | 'wrapWithTrans' | 'wrapWithT',
  [{ forceFix: string[] }]
>;

export const elementIsTrans = (node: TSESTree.Node) => {
  return (
    node.type === AST_NODE_TYPES.JSXElement &&
    node.openingElement.type === AST_NODE_TYPES.JSXOpeningElement &&
    node.openingElement.name.type === AST_NODE_TYPES.JSXIdentifier &&
    node.openingElement.name.name === 'Trans'
  );
};

const getParentMethod = (node: TSESTree.Node, context: TranslationRuleContext) => {
  const ancestors = context.sourceCode.getAncestors(node);
  return ancestors.find((anc) => {
    return (
      anc.type === AST_NODE_TYPES.ArrowFunctionExpression ||
      anc.type === AST_NODE_TYPES.FunctionDeclaration ||
      anc.type === AST_NODE_TYPES.FunctionExpression ||
      anc.type === AST_NODE_TYPES.ClassDeclaration
    );
  });
};

const isStringLiteral = (node: TSESTree.Node) => {
  return node.type === AST_NODE_TYPES.Literal && typeof node.value === 'string';
};

/**
 * Converts a string to kebab case
 */
function toKebabCase(str: string) {
  return str
    .replace(/([a-z])([A-Z])/g, '$1-$2')
    .toLowerCase()
    .replace(/\s+/g, '-');
}

/**
 * Checks if a string is non-alphanumeric
 */
export function isStringNonAlphanumeric(str: string) {
  return !/[a-zA-Z0-9]/.test(str);
}
/**
 * Checks if we _should_ fix an error automatically
 */
export function shouldBeFixed(context: TranslationRuleContext) {
  const pathsThatAreFixable = context.options[0]?.forceFix || [];
  return pathsThatAreFixable.some((path) => context.filename.includes(path));
}

/**
 * Checks if a node can be fixed automatically
 */
export function canBeFixed(
  node: TSESTree.JSXAttribute | TSESTree.JSXElement | TSESTree.JSXFragment | TSESTree.Property,
  context: TranslationRuleContext
) {
  if (!getTranslationPrefix(context)) {
    return false;
  }

  const parentMethod = getParentMethod(node, context);
  const isAttribute = node.type === AST_NODE_TYPES.JSXAttribute;
  const isProperty = node.type === AST_NODE_TYPES.Property;
  const isPropertyOrAttribute = isAttribute || isProperty;

  // We can only fix JSX attribute strings that are within a function,
  // otherwise the `t` function call will be made too early
  if (isPropertyOrAttribute && !parentMethod) {
    return false;
  }

  // If we're going to try and fix using `t`, and it already exists in the scope,
  // but not from `useTranslate`, then we can't fix/provide a suggestion
  if (isPropertyOrAttribute && parentMethod) {
    const hasTDeclaration = getTDeclaration(parentMethod, context);
    const hasUseTranslateDeclaration = methodHasUseTranslate(parentMethod, context);
    if (hasTDeclaration && !hasUseTranslateDeclaration) {
      return false;
    }
  }
  if (isAttribute && node.value?.type === AST_NODE_TYPES.JSXExpressionContainer) {
    return isStringLiteral(node.value.expression);
  }

  const values =
    node.type === AST_NODE_TYPES.JSXElement || node.type === AST_NODE_TYPES.JSXFragment
      ? node.children.map((child) => {
          return getNodeValue(child);
        })
      : [getNodeValue(node)];

  const stringIsTooLong = values.some((value) => value.trim().split(' ').length > 10);
  // If we have more than 10 words,
  // we don't want to fix it automatically as the chance of a duplicate key is higher,
  // and it's better for a user to manually decide the key
  if (stringIsTooLong) {
    return false;
  }
  const stringIsNonAlphanumeric = values.some((value) => !/[a-zA-Z0-9]/.test(value));
  const stringContainsHTMLEntities = values.some((value) => /(&[a-zA-Z0-9]+;)/.test(value));
  // If node only contains non-alphanumeric characters,
  // or contains HTML character entities, then we don't want to autofix
  if (stringIsNonAlphanumeric || stringContainsHTMLEntities) {
    return false;
  }

  return true;
}

/**
 * Gets the translation prefix from the filename
 */
export function getTranslationPrefix(context: TranslationRuleContext) {
  const filename = context.filename;
  const match = filename.match(/public\/app\/features\/(.+?)\//);
  if (match) {
    return match[1];
  }
  return null;
}

/**
 * Gets the i18n key for a node
 */
const getI18nKey = (
  node: TSESTree.JSXAttribute | TSESTree.JSXText | TSESTree.Property,
  context: TranslationRuleContext
) => {
  const prefixFromFilePath = getTranslationPrefix(context);
  const stringValue = getNodeValue(node);

  const componentNames = getComponentNames(node, context);

  const propertyName =
    node.type === AST_NODE_TYPES.Property && node.key.type === AST_NODE_TYPES.Identifier ? String(node.key.name) : null;

  const words = stringValue
    .trim()
    .replace(/[^\a-zA-Z\s]/g, '')
    .trim()
    .split(/\s+/);

  const maxWordsForKey = 6;

  // If we have more than 6 words, filter out the words that are less than 4 characters
  // This heuristic tends to result in a good balance between unique and descriptive keys
  const filteredWords = words.length > maxWordsForKey ? words.filter((word) => word.length > 4) : words;

  // If we've filtered everything out, use the original words, deduplicated
  const wordsToUse = filteredWords.length === 0 ? words : filteredWords;
  const uniqueWords = [...new Set(wordsToUse)].slice(0, maxWordsForKey);

  let kebabString = toKebabCase(uniqueWords.join(' '));

  if (node.type === AST_NODE_TYPES.JSXAttribute) {
    const propName = node.name.name;
    const attribute = node.parent?.attributes.find(
      (attr) =>
        attr.type === AST_NODE_TYPES.JSXAttribute &&
        attr.name.type === AST_NODE_TYPES.JSXIdentifier &&
        attr &&
        ['id', 'data-testid'].includes(attr.name?.name)
    );
    const potentialId =
      attribute &&
      attribute.type === AST_NODE_TYPES.JSXAttribute &&
      attribute.value &&
      attribute.value.type === AST_NODE_TYPES.Literal
        ? attribute.value.value
        : undefined;
    kebabString = [potentialId, propName, kebabString].filter(Boolean).join('-');
  }

  const fullPrefix = [prefixFromFilePath, ...componentNames, propertyName, kebabString].filter(Boolean).join('.');

  return fullPrefix;
};

/**
 * Gets component names from ancestors
 */
function getComponentNames(
  node: TSESTree.JSXAttribute | TSESTree.JSXText | TSESTree.Property,
  context: TranslationRuleContext
) {
  const names = [];
  const ancestors = context.sourceCode.getAncestors(node);

  for (const ancestor of ancestors) {
    if (
      ancestor.type === AST_NODE_TYPES.VariableDeclarator ||
      ancestor.type === AST_NODE_TYPES.FunctionDeclaration ||
      ancestor.type === AST_NODE_TYPES.ClassDeclaration
    ) {
      const name = ancestor.id?.type === AST_NODE_TYPES.Identifier ? ancestor.id.name : '';
      // Remove the word "component" from the name, as this is a bit
      // redundant in a translation key
      const sanitizedName = name.replace(/component/gi, '');
      names.push(toKebabCase(sanitizedName));
    }
  }

  return names;
}

/**
 * For a given node, check the scope and find a variable declaration of `t`
 */
function getTDeclaration(node: TSESTree.Node, context: TranslationRuleContext) {
  return context.sourceCode.getScope(node).variables.find((v) => v.name === 't');
}

/**
 * Checks if a node has a variable declaration of `t`
 * that came from a `useTranslate` call
 */
function methodHasUseTranslate(node: TSESTree.Node, context: TranslationRuleContext) {
  const tDeclaration = getTDeclaration(node, context);
  return (
    tDeclaration &&
    tDeclaration.defs.find((definition) => {
      const isVariableDeclaration = definition.node.type === AST_NODE_TYPES.VariableDeclarator;
      const declarationInit = isVariableDeclaration ? definition.node.init : null;
      return (
        isVariableDeclaration &&
        declarationInit &&
        declarationInit.type === AST_NODE_TYPES.CallExpression &&
        declarationInit.callee.type === AST_NODE_TYPES.Identifier &&
        declarationInit.callee.name === 'useTranslate'
      );
    })
  );
}

/**
 * Gets the import fixer for a node
 */
function getImportsFixer(
  node: TSESTree.JSXElement | TSESTree.JSXFragment | TSESTree.JSXAttribute | TSESTree.Property,
  fixer: RuleFixer,
  importName: 'Trans' | 't' | 'useTranslate',
  context: TranslationRuleContext
) {
  const body = context.sourceCode.ast.body;

  /** Map of where we expect to import each translation util from */
  const importPackage = {
    Trans: '@grafana/i18n',
    useTranslate: '@grafana/i18n',
    t: '@grafana/i18n/internal',
  };

  const parentMethod = getParentMethod(node, context);

  if (importName === 't') {
    // If we're trying to import `t`,
    // and there's already a `t` variable declaration in the parent method that came from `useTranslate`,
    // do nothing
    const declarationFromUseTranslate = parentMethod ? methodHasUseTranslate(parentMethod, context) : false;
    if (declarationFromUseTranslate) {
      return;
    }
  }

  const expectedImport = importPackage[importName];

  const existingAppCoreI18n = body.find(
    (node) => node.type === AST_NODE_TYPES.ImportDeclaration && node.source.value === importPackage[importName]
  );

  // If there's no existing import at all, add it
  if (!existingAppCoreI18n) {
    return fixer.insertTextBefore(body[0], `import { ${importName} } from '${expectedImport}';\n`);
  }

  // To keep the typechecker happy - we have to explicitly check the type
  // so we can infer it further down
  if (existingAppCoreI18n.type !== AST_NODE_TYPES.ImportDeclaration) {
    return;
  }

  // If there's an existing import, and it already has the importName, do nothing
  if (
    existingAppCoreI18n.specifiers.some((s) => {
      return (
        s.type === AST_NODE_TYPES.ImportSpecifier &&
        s.imported.type === AST_NODE_TYPES.Identifier &&
        s.imported.name === importName
      );
    })
  ) {
    return;
  }
  const lastSpecifier = existingAppCoreI18n.specifiers[existingAppCoreI18n.specifiers.length - 1];
  const range = [lastSpecifier.range[1], lastSpecifier.range[1]] as const;
  return fixer.insertTextAfterRange(range, `, ${importName}`);
}

export const getTransFixers =
  (node: TSESTree.JSXElement | TSESTree.JSXFragment, context: TranslationRuleContext) => (fixer: RuleFixer) => {
    const fixes = [];
    const children = node.children;
    children.forEach((child) => {
      if (child.type === AST_NODE_TYPES.JSXText) {
        const i18nKey = getI18nKey(child, context);
        const value = getNodeValue(child);
        fixes.push(fixer.replaceText(child, `<Trans i18nKey="${i18nKey}">${value}</Trans>`));
      }
    });

    const importsFixer = getImportsFixer(node, fixer, 'Trans', context);
    if (importsFixer) {
      fixes.push(importsFixer);
    }
    return fixes;
  };

const firstCharIsUpper = (str: string) => {
  return str.charAt(0) === str.charAt(0).toUpperCase();
};

const getUseTranslateFixer = (
  node: TSESTree.JSXAttribute | TSESTree.Property,
  fixer: RuleFixer,
  context: TranslationRuleContext
) => {
  const parentMethod = getParentMethod(node, context);

  const functionIsNotUpperCase =
    parentMethod &&
    parentMethod.type === AST_NODE_TYPES.FunctionDeclaration &&
    (!parentMethod.id || !firstCharIsUpper(parentMethod.id.name));

  const variableDeclaratorIsNotUpperCase =
    parentMethod &&
    parentMethod.parent.type === AST_NODE_TYPES.VariableDeclarator &&
    parentMethod.parent.id.type === AST_NODE_TYPES.Identifier &&
    !firstCharIsUpper(parentMethod.parent.id.name);

  // If the node is not within a function, or the parent method does not start with an uppercase letter,
  // then we can't reliably add `useTranslate`, as this may not be a React component
  if (
    !parentMethod ||
    functionIsNotUpperCase ||
    variableDeclaratorIsNotUpperCase ||
    parentMethod.body.type !== AST_NODE_TYPES.BlockStatement
  ) {
    return;
  }

  const returnStatement = parentMethod.body.body.find((node) => node.type === AST_NODE_TYPES.ReturnStatement);
  if (!returnStatement) {
    return;
  }

  const returnStatementIsJsx =
    returnStatement.argument &&
    (returnStatement.argument.type === AST_NODE_TYPES.JSXElement ||
      returnStatement.argument.type === AST_NODE_TYPES.JSXFragment);

  if (!returnStatementIsJsx) {
    return;
  }
  const tDeclarationExists = getTDeclaration(parentMethod, context);
  const useTranslateExists = methodHasUseTranslate(parentMethod, context);

  if (tDeclarationExists && useTranslateExists) {
    return;
  }

  // If we've got all this way, then:
  // - There is a parent method
  // - It returns JSX
  // - The method name starts with a capital letter
  // - There is not already a call to `useTranslate` in the parent method
  // In that scenario, we assume that we can fix and add a usage of the hook to the start of the body of the method
  return fixer.insertTextBefore(parentMethod.body.body[0], 'const { t } = useTranslate();\n');
};

export const getTFixers =
  (node: TSESTree.JSXAttribute | TSESTree.Property, context: TranslationRuleContext) => (fixer: RuleFixer) => {
    const fixes = [];
    const i18nKey = getI18nKey(node, context);
    const value = getNodeValue(node);
    const wrappingQuotes = value.includes('"') ? "'" : '"';

    if (node.type === AST_NODE_TYPES.Property) {
      fixes.push(fixer.replaceText(node.value, `t("${i18nKey}", ${wrappingQuotes}${value}${wrappingQuotes})`));
    } else {
      fixes.push(
        fixer.replaceText(node, `${node.name.name}={t("${i18nKey}", ${wrappingQuotes}${value}${wrappingQuotes})}`)
      );
    }

    // Check if we need to add `useTranslate` to the node
    const useTranslateFixer = getUseTranslateFixer(node, fixer, context);
    if (useTranslateFixer) {
      fixes.push(useTranslateFixer);
    }

    // Check if we need to add `t` or `useTranslate` to the imports
    const importToAdd = useTranslateFixer ? 'useTranslate' : 't';
    const importsFixer = getImportsFixer(node, fixer, importToAdd, context);
    if (importsFixer) {
      fixes.push(importsFixer);
    }

    return fixes;
  };

/**
 * Gets the value of a node
 */
export function getNodeValue(
  node:
    | TSESTree.JSXAttribute
    | TSESTree.JSXText
    | TSESTree.JSXElement
    | TSESTree.JSXFragment
    | TSESTree.JSXChild
    | TSESTree.Property
) {
  if (
    (node.type === AST_NODE_TYPES.JSXAttribute || node.type === AST_NODE_TYPES.Property) &&
    node.value?.type === AST_NODE_TYPES.Literal
  ) {
    // TODO: Update this to return bool/number values and handle the type issues elsewhere
    // For now, we'll just return an empty string so we consider any numbers or booleans as not being issues
    if (typeof node.value.value === 'boolean' || typeof node.value.value === 'number') {
      return '';
    }
    return String(node.value.value) || '';
  }
  if (node.type === AST_NODE_TYPES.JSXText) {
    // Return the raw value if we can, so we can work out if there are any HTML entities
    return node.raw;
  }
  if (node.type === AST_NODE_TYPES.JSXAttribute && node.value?.type === AST_NODE_TYPES.JSXExpressionContainer) {
    // this condition is basically `isStringLiteral`, but we can't use the function
    // else it doesn't narrow the type correctly :(
    if (node.value.expression.type === AST_NODE_TYPES.Literal && typeof node.value.expression.value === 'string') {
      return node.value.expression.value;
    }
  }
  return '';
}
