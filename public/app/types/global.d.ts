declare module 'combokeys' {
  class Combokeys {
    constructor(element?: HTMLElement);
    bind(keys: string | string[], callback: any): void;
    unbind(keys: string | string[]): void;
    reset(): void;
  }
  export default Combokeys;
}

declare module '*/search.terms' {
  const terms: any;
  export = terms;
}

declare module '*/search' {
  export const parser: any;
}

declare module '*/jsonMarkup' {
  const jsonMarkup: any;
  export default jsonMarkup;
}
