/**
 * FilterEngine Tests
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { FilterEngine, debounce } from '../src/utils/filter-engine';
import type { Row, Column, Filter, FilterState, RowData } from '../src/types';

interface TestData extends RowData {
  id: number;
  name: string;
  age: number;
  email: string;
  active: boolean;
  score: number | null;
  joinDate: string;
}

describe('FilterEngine', () => {
  let filterEngine: FilterEngine<TestData>;
  let columns: Column<TestData>[];
  let rows: Row<TestData>[];

  beforeEach(() => {
    const options = {
      filtering: {
        enabled: true,
        globalSearch: true,
        debounceMs: 300,
        caseSensitive: false,
        customFilters: {}
      }
    };

    filterEngine = new FilterEngine(options);

    columns = [
      { id: 'id', field: 'id', label: 'ID' },
      { id: 'name', field: 'name', label: 'Name' },
      { id: 'age', field: 'age', label: 'Age' },
      { id: 'email', field: 'email', label: 'Email' },
      { id: 'active', field: 'active', label: 'Active' },
      { id: 'score', field: 'score', label: 'Score' },
      { id: 'joinDate', field: 'joinDate', label: 'Join Date' },
    ];

    rows = [
      {
        id: 1,
        data: { id: 1, name: 'John Doe', age: 30, email: 'john@example.com', active: true, score: 85, joinDate: '2023-01-15' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 2,
        data: { id: 2, name: 'Jane Smith', age: 25, email: 'jane@example.com', active: false, score: 92, joinDate: '2023-02-20' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 3,
        data: { id: 3, name: 'Bob Johnson', age: 35, email: 'bob@company.org', active: true, score: null, joinDate: '2022-12-10' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 4,
        data: { id: 4, name: 'Alice Brown', age: 28, email: 'alice@example.com', active: false, score: 78, joinDate: '2023-03-05' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 5,
        data: { id: 5, name: 'Charlie Wilson', age: 42, email: 'charlie@example.com', active: true, score: 90, joinDate: '2022-11-01' },
        selected: false,
        expanded: false,
        disabled: false
      }
    ];
  });

  describe('Basic Filtering', () => {
    it('should return all rows when no filters applied', () => {
      const filterState: FilterState = { filters: [], globalSearch: '' };
      const result = filterEngine.filter(rows, filterState, columns);
      expect(result).toHaveLength(5);
      expect(result).toEqual(rows);
    });

    it('should filter rows with empty filter state', () => {
      const filterState: FilterState = { filters: [], globalSearch: '' };
      const result = filterEngine.filter(rows, filterState, columns);
      expect(result).toEqual(rows);
    });
  });

  describe('Column Filters', () => {
    describe('Equals Operator', () => {
      it('should filter by equals operator', () => {
        const filters: Filter[] = [
          { columnId: 'active', operator: 'equals', value: true }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(3);
        expect(result.every(row => row.data.active === true)).toBe(true);
      });

      it('should filter by equals with string value', () => {
        const filters: Filter[] = [
          { columnId: 'name', operator: 'equals', value: 'John Doe' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(1);
        expect(result[0].data.name).toBe('John Doe');
      });
    });

    describe('Not Equals Operator', () => {
      it('should filter by not equals operator', () => {
        const filters: Filter[] = [
          { columnId: 'active', operator: 'notEquals', value: true }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(2);
        expect(result.every(row => row.data.active === false)).toBe(true);
      });
    });

    describe('Contains Operator', () => {
      it('should filter by contains operator', () => {
        const filters: Filter[] = [
          { columnId: 'name', operator: 'contains', value: 'john' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(2); // John Doe and Bob Johnson
        expect(result.every(row => 
          row.data.name.toLowerCase().includes('john')
        )).toBe(true);
      });

      it('should filter by contains with case sensitivity', () => {
        const caseSensitiveEngine = new FilterEngine({
          filtering: {
            enabled: true,
            globalSearch: true,
            debounceMs: 300,
            caseSensitive: true,
            customFilters: {}
          }
        });

        const filters: Filter[] = [
          { columnId: 'name', operator: 'contains', value: 'Doe' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = caseSensitiveEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(1); // Only John Doe (exact case "John")
        expect(result[0].data.name).toBe('John Doe');
      });
    });

    describe('Not Contains Operator', () => {
      it('should filter by not contains operator', () => {
        const filters: Filter[] = [
          { columnId: 'email', operator: 'notContains', value: 'example.com' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(1); // Only Bob with company.org
        expect(result[0].data.email).toBe('bob@company.org');
      });
    });

    describe('Starts With Operator', () => {
      it('should filter by starts with operator', () => {
        const filters: Filter[] = [
          { columnId: 'name', operator: 'startsWith', value: 'a' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(1); // Alice Brown
        expect(result[0].data.name).toBe('Alice Brown');
      });
    });

    describe('Ends With Operator', () => {
      it('should filter by ends with operator', () => {
        const filters: Filter[] = [
          { columnId: 'email', operator: 'endsWith', value: 'company.org' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(1); // Bob Johnson
        expect(result[0].data.email).toBe('bob@company.org');
      });
    });

    describe('Empty/Not Empty Operators', () => {
      it('should filter by isEmpty operator', () => {
        const filters: Filter[] = [
          { columnId: 'score', operator: 'isEmpty', value: null }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(1); // Bob Johnson with null score
        expect(result[0].data.score).toBeNull();
      });

      it('should filter by isNotEmpty operator', () => {
        const filters: Filter[] = [
          { columnId: 'score', operator: 'isNotEmpty', value: null }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(4); // All except Bob Johnson
        expect(result.every(row => row.data.score !== null)).toBe(true);
      });
    });

    describe('Numeric Comparison Operators', () => {
      it('should filter by greater than operator', () => {
        const filters: Filter[] = [
          { columnId: 'age', operator: 'greaterThan', value: 30 }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(2); // Bob (35) and Charlie (42)
        expect(result.every(row => row.data.age > 30)).toBe(true);
      });

      it('should filter by greater than or equal operator', () => {
        const filters: Filter[] = [
          { columnId: 'age', operator: 'greaterThanOrEqual', value: 30 }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(3); // John (30), Bob (35), Charlie (42)
        expect(result.every(row => row.data.age >= 30)).toBe(true);
      });

      it('should filter by less than operator', () => {
        const filters: Filter[] = [
          { columnId: 'age', operator: 'lessThan', value: 30 }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(2); // Jane (25) and Alice (28)
        expect(result.every(row => row.data.age < 30)).toBe(true);
      });

      it('should filter by less than or equal operator', () => {
        const filters: Filter[] = [
          { columnId: 'age', operator: 'lessThanOrEqual', value: 30 }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(3); // Jane (25), Alice (28), John (30)
        expect(result.every(row => row.data.age <= 30)).toBe(true);
      });
    });

    describe('Between Operator', () => {
      it('should filter by between operator', () => {
        const filters: Filter[] = [
          { columnId: 'age', operator: 'between', value: [25, 35] }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(4); // Jane (25), Alice (28), John (30), Bob (35)
        expect(result.every(row => row.data.age >= 25 && row.data.age <= 35)).toBe(true);
      });

      it('should handle invalid between values gracefully', () => {
        const filters: Filter[] = [
          { columnId: 'age', operator: 'between', value: [25] } // Invalid - not array of 2
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(0); // Should return empty when invalid
      });
    });

    describe('In/Not In Operators', () => {
      it('should filter by in operator', () => {
        const filters: Filter[] = [
          { columnId: 'name', operator: 'in', value: ['john doe', 'jane smith'] }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(2); // John and Jane
        expect(result.map(r => r.data.name.toLowerCase())).toEqual(['john doe', 'jane smith']);
      });

      it('should filter by not in operator', () => {
        const filters: Filter[] = [
          { columnId: 'name', operator: 'notIn', value: ['john doe', 'jane smith'] }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(3); // Bob, Alice, Charlie
        expect(result.every(row => 
          !['john doe', 'jane smith'].includes(row.data.name.toLowerCase())
        )).toBe(true);
      });

      it('should handle non-array values for in/notIn gracefully', () => {
        const filters: Filter[] = [
          { columnId: 'name', operator: 'in', value: 'john doe' } // Should be array
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(0); // Should return empty for invalid format
      });
    });

    describe('Multiple Column Filters', () => {
      it('should apply multiple filters (AND logic)', () => {
        const filters: Filter[] = [
          { columnId: 'active', operator: 'equals', value: true },
          { columnId: 'age', operator: 'greaterThan', value: 30 }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toHaveLength(2); // Bob (35, active) and Charlie (42, active)
        expect(result.every(row => row.data.active === true && row.data.age > 30)).toBe(true);
      });

      it('should handle filters on non-existent columns', () => {
        const filters: Filter[] = [
          { columnId: 'nonexistent', operator: 'equals', value: 'test' }
        ];
        const filterState: FilterState = { filters, globalSearch: '' };
        const result = filterEngine.filter(rows, filterState, columns);
        
        expect(result).toEqual(rows); // Should return all rows when column not found
      });
    });
  });

  describe('Global Search', () => {
    it('should perform global search across all columns', () => {
      const filterState: FilterState = { 
        filters: [], 
        globalSearch: 'john' 
      };
      const result = filterEngine.filter(rows, filterState, columns);
      
      expect(result).toHaveLength(2); // John Doe and Bob Johnson
      expect(result.some(row => row.data.name.toLowerCase().includes('john'))).toBe(true);
    });

    it('should search in email field', () => {
      const filterState: FilterState = { 
        filters: [], 
        globalSearch: 'company.org' 
      };
      const result = filterEngine.filter(rows, filterState, columns);
      
      expect(result).toHaveLength(1); // Bob Johnson
      expect(result[0].data.email).toBe('bob@company.org');
    });

    it('should ignore whitespace-only search', () => {
      const filterState: FilterState = { 
        filters: [], 
        globalSearch: '   ' 
      };
      const result = filterEngine.filter(rows, filterState, columns);
      
      expect(result).toEqual(rows); // Should return all rows
    });

    it('should respect filterable column setting', () => {
      const nonFilterableColumns = columns.map(col => 
        col.id === 'email' ? { ...col, filterable: false } : col
      );
      
      const filterState: FilterState = { 
        filters: [], 
        globalSearch: 'company.org' 
      };
      const result = filterEngine.filter(rows, filterState, nonFilterableColumns);
      
      expect(result).toHaveLength(0); // Should not find email content
    });
  });

  describe('Combined Filtering', () => {
    it('should apply both column filters and global search', () => {
      const filters: Filter[] = [
        { columnId: 'active', operator: 'equals', value: true }
      ];
      const filterState: FilterState = { 
        filters, 
        globalSearch: 'charlie' 
      };
      const result = filterEngine.filter(rows, filterState, columns);
      
      expect(result).toHaveLength(1); // Only Charlie (active and name contains charlie)
      expect(result[0].data.name).toBe('Charlie Wilson');
      expect(result[0].data.active).toBe(true);
    });
  });

  describe('Custom Filters', () => {
    it('should use custom filter function when provided', () => {
      const customFilterEngine = new FilterEngine({
        filtering: {
          enabled: true,
          globalSearch: true,
          debounceMs: 300,
          caseSensitive: false,
          customFilters: {
            'age': (row, filter) => {
              // Custom logic: filter for even ages only
              return row.data.age % 2 === 0;
            }
          }
        }
      });

      const filters: Filter[] = [
        { columnId: 'age', operator: 'equals', value: 'any' } // Value doesn't matter for custom filter
      ];
      const filterState: FilterState = { filters, globalSearch: '' };
      const result = customFilterEngine.filter(rows, filterState, columns);
      
      expect(result).toHaveLength(3); // John (30), Alice (28), and Charlie (42) - even ages
      expect(result.every(row => row.data.age % 2 === 0)).toBe(true);
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty rows array', () => {
      const filterState: FilterState = { 
        filters: [{ columnId: 'name', operator: 'contains', value: 'test' }], 
        globalSearch: 'search' 
      };
      const result = filterEngine.filter([], filterState, columns);
      
      expect(result).toEqual([]);
    });

    it('should handle empty columns array', () => {
      const filterState: FilterState = { 
        filters: [], 
        globalSearch: 'search' 
      };
      const result = filterEngine.filter(rows, filterState, []);
      
      expect(result).toEqual(rows); // Global search with no searchable columns returns all rows
    });

    it('should handle unknown filter operators gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      const filters: Filter[] = [
        { columnId: 'name', operator: 'unknownOperator' as any, value: 'test' }
      ];
      const filterState: FilterState = { filters, globalSearch: '' };
      const result = filterEngine.filter(rows, filterState, columns);
      
      expect(result).toEqual(rows); // Should return all rows for unknown operator
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('FilterEngine: Unknown operator unknownOperator')
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle columns with custom accessors', () => {
      const customColumns = [
        ...columns,
        {
          id: 'fullInfo',
          field: 'name' as keyof TestData,
          label: 'Full Info',
          accessor: (row: TestData) => `${row.name} (${row.age})`
        }
      ];

      const filters: Filter[] = [
        { columnId: 'fullInfo', operator: 'contains', value: '30' }
      ];
      const filterState: FilterState = { filters, globalSearch: '' };
      const result = filterEngine.filter(rows, filterState, customColumns);
      
      expect(result).toHaveLength(1); // John Doe (30)
      expect(result[0].data.name).toBe('John Doe');
      expect(result[0].data.age).toBe(30);
    });
  });
});

describe('Debounce Utility', () => {
  it('should debounce function calls', async () => {
    const fn = vi.fn();
    const debouncedFn = debounce(fn, 100);
    
    // Call multiple times rapidly
    debouncedFn('call1');
    debouncedFn('call2');
    debouncedFn('call3');
    
    // Should not be called yet
    expect(fn).not.toHaveBeenCalled();
    
    // Wait for debounce delay
    await new Promise(resolve => setTimeout(resolve, 150));
    
    // Should be called once with last argument
    expect(fn).toHaveBeenCalledTimes(1);
    expect(fn).toHaveBeenCalledWith('call3');
  });

  it('should cancel previous calls when new call is made', async () => {
    const fn = vi.fn();
    const debouncedFn = debounce(fn, 100);
    
    debouncedFn('first');
    
    // Wait halfway
    await new Promise(resolve => setTimeout(resolve, 50));
    
    debouncedFn('second'); // This should cancel the first call
    
    // Wait for full delay
    await new Promise(resolve => setTimeout(resolve, 150));
    
    expect(fn).toHaveBeenCalledTimes(1);
    expect(fn).toHaveBeenCalledWith('second');
  });

  it('should handle multiple arguments', async () => {
    const fn = vi.fn();
    const debouncedFn = debounce(fn, 50);
    
    debouncedFn('arg1', 'arg2', 'arg3');
    
    await new Promise(resolve => setTimeout(resolve, 100));
    
    expect(fn).toHaveBeenCalledWith('arg1', 'arg2', 'arg3');
  });
});