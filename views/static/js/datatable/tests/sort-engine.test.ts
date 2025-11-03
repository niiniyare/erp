/**
 * SortEngine Tests
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { SortEngine } from '../src/utils/sort-engine';
import type { Row, Column, SortState, MultiSortState, RowData, CellValue } from '../src/types';

interface TestData extends RowData {
  id: number;
  name: string;
  age: number;
  email: string;
  active: boolean;
  score: number | null;
  joinDate: string;
  priority: string;
}

describe('SortEngine', () => {
  let sortEngine: SortEngine<TestData>;
  let columns: Column<TestData>[];
  let rows: Row<TestData>[];

  beforeEach(() => {
    const options = {
      sorting: {
        enabled: true,
        multiSort: false,
        defaultSort: undefined
      },
      customComparators: {}
    };

    sortEngine = new SortEngine(options);

    columns = [
      { id: 'id', field: 'id', label: 'ID' },
      { id: 'name', field: 'name', label: 'Name' },
      { id: 'age', field: 'age', label: 'Age' },
      { id: 'email', field: 'email', label: 'Email' },
      { id: 'active', field: 'active', label: 'Active' },
      { id: 'score', field: 'score', label: 'Score' },
      { id: 'joinDate', field: 'joinDate', label: 'Join Date' },
      { id: 'priority', field: 'priority', label: 'Priority' },
    ];

    rows = [
      {
        id: 1,
        data: { id: 1, name: 'John Doe', age: 30, email: 'john@example.com', active: true, score: 85, joinDate: '2023-01-15', priority: 'high' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 2,
        data: { id: 2, name: 'Jane Smith', age: 25, email: 'jane@example.com', active: false, score: 92, joinDate: '2023-02-20', priority: 'medium' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 3,
        data: { id: 3, name: 'Bob Johnson', age: 35, email: 'bob@company.org', active: true, score: null, joinDate: '2022-12-10', priority: 'low' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 4,
        data: { id: 4, name: 'Alice Brown', age: 28, email: 'alice@example.com', active: false, score: 78, joinDate: '2023-03-05', priority: 'high' },
        selected: false,
        expanded: false,
        disabled: false
      },
      {
        id: 5,
        data: { id: 5, name: 'Charlie Wilson', age: 42, email: 'charlie@example.com', active: true, score: 90, joinDate: '2022-11-01', priority: 'medium' },
        selected: false,
        expanded: false,
        disabled: false
      }
    ];
  });

  describe('Single Column Sorting', () => {
    describe('String Sorting', () => {
      it('should sort by name ascending', () => {
        const sortState: SortState = { columnId: 'name', direction: 'asc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const names = result.map(row => row.data.name);
        expect(names).toEqual(['Alice Brown', 'Bob Johnson', 'Charlie Wilson', 'Jane Smith', 'John Doe']);
      });

      it('should sort by name descending', () => {
        const sortState: SortState = { columnId: 'name', direction: 'desc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const names = result.map(row => row.data.name);
        expect(names).toEqual(['John Doe', 'Jane Smith', 'Charlie Wilson', 'Bob Johnson', 'Alice Brown']);
      });

      it('should handle case-insensitive string sorting', () => {
        const mixedCaseRows = [
          { ...rows[0], data: { ...rows[0].data, name: 'alice' } },
          { ...rows[1], data: { ...rows[1].data, name: 'Bob' } },
          { ...rows[2], data: { ...rows[2].data, name: 'CHARLIE' } },
        ];

        const sortState: SortState = { columnId: 'name', direction: 'asc' };
        const result = sortEngine.sort(mixedCaseRows, sortState, columns);
        
        const names = result.map(row => row.data.name);
        expect(names).toEqual(['alice', 'Bob', 'CHARLIE']);
      });
    });

    describe('Number Sorting', () => {
      it('should sort by age ascending', () => {
        const sortState: SortState = { columnId: 'age', direction: 'asc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const ages = result.map(row => row.data.age);
        expect(ages).toEqual([25, 28, 30, 35, 42]);
      });

      it('should sort by age descending', () => {
        const sortState: SortState = { columnId: 'age', direction: 'desc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const ages = result.map(row => row.data.age);
        expect(ages).toEqual([42, 35, 30, 28, 25]);
      });

      it('should handle null values in numeric sorting', () => {
        const sortState: SortState = { columnId: 'score', direction: 'asc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        // Null should come first (treated as less than any number)
        expect(result[0].data.score).toBeNull();
        expect(result[0].data.name).toBe('Bob Johnson');
        
        // Then sorted numbers
        const scoresWithoutNull = result.slice(1).map(row => row.data.score);
        expect(scoresWithoutNull).toEqual([78, 85, 90, 92]);
      });

      it('should handle null values in numeric sorting descending', () => {
        const sortState: SortState = { columnId: 'score', direction: 'desc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        // Numbers first (highest to lowest)
        const firstFour = result.slice(0, 4).map(row => row.data.score);
        expect(firstFour).toEqual([92, 90, 85, 78]);
        
        // Null last
        expect(result[4].data.score).toBeNull();
      });
    });

    describe('Boolean Sorting', () => {
      it('should sort by active status ascending (false first)', () => {
        const sortState: SortState = { columnId: 'active', direction: 'asc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const activeStatuses = result.map(row => row.data.active);
        expect(activeStatuses).toEqual([false, false, true, true, true]);
      });

      it('should sort by active status descending (true first)', () => {
        const sortState: SortState = { columnId: 'active', direction: 'desc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const activeStatuses = result.map(row => row.data.active);
        expect(activeStatuses).toEqual([true, true, true, false, false]);
      });
    });

    describe('Date Sorting', () => {
      it('should sort by join date ascending', () => {
        const sortState: SortState = { columnId: 'joinDate', direction: 'asc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const dates = result.map(row => row.data.joinDate);
        expect(dates).toEqual(['2022-11-01', '2022-12-10', '2023-01-15', '2023-02-20', '2023-03-05']);
      });

      it('should sort by join date descending', () => {
        const sortState: SortState = { columnId: 'joinDate', direction: 'desc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        const dates = result.map(row => row.data.joinDate);
        expect(dates).toEqual(['2023-03-05', '2023-02-20', '2023-01-15', '2022-12-10', '2022-11-01']);
      });

      it('should handle Date objects', () => {
        const dateRows = rows.map(row => ({
          ...row,
          data: { ...row.data, joinDate: new Date(row.data.joinDate) as any }
        }));

        const sortState: SortState = { columnId: 'joinDate', direction: 'asc' };
        const result = sortEngine.sort(dateRows, sortState, columns);
        
        const dates = result.map(row => (row.data.joinDate as Date).toISOString().substring(0, 10));
        expect(dates).toEqual(['2022-11-01', '2022-12-10', '2023-01-15', '2023-02-20', '2023-03-05']);
      });
    });

    describe('String with Numbers (Natural Sorting)', () => {
      it('should handle natural sorting for strings with numbers', () => {
        const naturalRows = [
          { ...rows[0], data: { ...rows[0].data, priority: 'item1' } },
          { ...rows[1], data: { ...rows[1].data, priority: 'item10' } },
          { ...rows[2], data: { ...rows[2].data, priority: 'item2' } },
          { ...rows[3], data: { ...rows[3].data, priority: 'item20' } },
          { ...rows[4], data: { ...rows[4].data, priority: 'item3' } },
        ];

        const sortState: SortState = { columnId: 'priority', direction: 'asc' };
        const result = sortEngine.sort(naturalRows, sortState, columns);
        
        const priorities = result.map(row => row.data.priority);
        expect(priorities).toEqual(['item1', 'item2', 'item3', 'item10', 'item20']);
      });
    });

    describe('Edge Cases', () => {
      it('should return original array when no sort direction', () => {
        const sortState: SortState = { columnId: 'name', direction: null };
        const result = sortEngine.sort(rows, sortState, columns);
        
        expect(result).toEqual(rows);
      });

      it('should return original array when column not found', () => {
        const sortState: SortState = { columnId: 'nonexistent', direction: 'asc' };
        const result = sortEngine.sort(rows, sortState, columns);
        
        expect(result).toEqual(rows);
      });

      it('should not mutate original array', () => {
        const originalOrder = rows.map(row => row.data.name);
        const sortState: SortState = { columnId: 'name', direction: 'asc' };
        
        sortEngine.sort(rows, sortState, columns);
        
        // Original array should remain unchanged
        const currentOrder = rows.map(row => row.data.name);
        expect(currentOrder).toEqual(originalOrder);
      });

      it('should handle empty array', () => {
        const sortState: SortState = { columnId: 'name', direction: 'asc' };
        const result = sortEngine.sort([], sortState, columns);
        
        expect(result).toEqual([]);
      });
    });
  });

  describe('Multi-Column Sorting', () => {
    it('should sort by multiple columns', () => {
      const multiSortState: MultiSortState = {
        sorts: [
          { columnId: 'active', direction: 'desc' }, // Active first
          { columnId: 'age', direction: 'asc' }      // Then by age ascending
        ]
      };
      
      const result = sortEngine.sort(rows, multiSortState, columns);
      
      // First should be active users by age: John (30), Bob (35), Charlie (42)
      // Then inactive users by age: Jane (25), Alice (28)
      const expectedOrder = [
        'John Doe',     // active: true, age: 30
        'Bob Johnson',  // active: true, age: 35
        'Charlie Wilson', // active: true, age: 42
        'Jane Smith',   // active: false, age: 25
        'Alice Brown'   // active: false, age: 28
      ];
      
      const actualOrder = result.map(row => row.data.name);
      expect(actualOrder).toEqual(expectedOrder);
    });

    it('should handle multi-sort with null directions', () => {
      const multiSortState: MultiSortState = {
        sorts: [
          { columnId: 'active', direction: null },    // Skip this one
          { columnId: 'age', direction: 'asc' }       // Sort by this
        ]
      };
      
      const result = sortEngine.sort(rows, multiSortState, columns);
      
      // Should only sort by age since first sort has null direction
      const ages = result.map(row => row.data.age);
      expect(ages).toEqual([25, 28, 30, 35, 42]);
    });

    it('should handle multi-sort with non-existent columns', () => {
      const multiSortState: MultiSortState = {
        sorts: [
          { columnId: 'nonexistent', direction: 'asc' }, // Skip this one
          { columnId: 'age', direction: 'asc' }          // Sort by this
        ]
      };
      
      const result = sortEngine.sort(rows, multiSortState, columns);
      
      // Should only sort by age since first column doesn't exist
      const ages = result.map(row => row.data.age);
      expect(ages).toEqual([25, 28, 30, 35, 42]);
    });

    it('should handle three-level sorting', () => {
      // Add some duplicate data to test three-level sorting
      const extendedRows = [
        ...rows,
        {
          id: 6,
          data: { id: 6, name: 'David Wilson', age: 30, email: 'david@example.com', active: true, score: 85, joinDate: '2023-04-01', priority: 'high' },
          selected: false,
          expanded: false,
          disabled: false
        }
      ];

      const multiSortState: MultiSortState = {
        sorts: [
          { columnId: 'active', direction: 'desc' }, // Active first
          { columnId: 'age', direction: 'asc' },     // Then by age
          { columnId: 'name', direction: 'asc' }     // Then by name
        ]
      };
      
      const result = sortEngine.sort(extendedRows, multiSortState, columns);
      
      // Both John and David are active with age 30, so should be sorted by name
      const activeAge30 = result.filter(row => row.data.active && row.data.age === 30);
      expect(activeAge30.map(row => row.data.name)).toEqual(['David Wilson', 'John Doe']);
    });
  });

  describe('Custom Comparators', () => {
    it('should use column-specific custom comparator', () => {
      const customColumns = columns.map(col => 
        col.id === 'priority' ? {
          ...col,
          comparator: (a: CellValue, b: CellValue): number => {
            const priorityOrder = { high: 3, medium: 2, low: 1 };
            const aVal = priorityOrder[a as keyof typeof priorityOrder] || 0;
            const bVal = priorityOrder[b as keyof typeof priorityOrder] || 0;
            return aVal - bVal;
          }
        } : col
      );

      const sortState: SortState = { columnId: 'priority', direction: 'asc' };
      const result = sortEngine.sort(rows, sortState, customColumns);
      
      const priorities = result.map(row => row.data.priority);
      expect(priorities).toEqual(['low', 'medium', 'medium', 'high', 'high']);
    });

    it('should use global custom comparator', () => {
      const customSortEngine = new SortEngine({
        sorting: {
          enabled: true,
          multiSort: false,
          defaultSort: undefined
        },
        customComparators: {
          'priority': (a: CellValue, b: CellValue): number => {
            const priorityOrder = { high: 3, medium: 2, low: 1 };
            const aVal = priorityOrder[a as keyof typeof priorityOrder] || 0;
            const bVal = priorityOrder[b as keyof typeof priorityOrder] || 0;
            return aVal - bVal;
          }
        }
      });

      const sortState: SortState = { columnId: 'priority', direction: 'desc' };
      const result = customSortEngine.sort(rows, sortState, columns);
      
      const priorities = result.map(row => row.data.priority);
      expect(priorities).toEqual(['high', 'high', 'medium', 'medium', 'low']);
    });

    it('should prefer column comparator over global comparator', () => {
      const customSortEngine = new SortEngine({
        sorting: {
          enabled: true,
          multiSort: false,
          defaultSort: undefined
        },
        customComparators: {
          'priority': (): number => 0 // Global comparator that makes everything equal
        }
      });

      const customColumns = columns.map(col => 
        col.id === 'priority' ? {
          ...col,
          comparator: (a: CellValue, b: CellValue): number => {
            // Column comparator that actually sorts
            const priorityOrder = { high: 3, medium: 2, low: 1 };
            const aVal = priorityOrder[a as keyof typeof priorityOrder] || 0;
            const bVal = priorityOrder[b as keyof typeof priorityOrder] || 0;
            return aVal - bVal;
          }
        } : col
      );

      const sortState: SortState = { columnId: 'priority', direction: 'asc' };
      const result = customSortEngine.sort(rows, sortState, customColumns);
      
      // Should use column comparator, not global one
      const priorities = result.map(row => row.data.priority);
      expect(priorities).toEqual(['low', 'medium', 'medium', 'high', 'high']);
    });
  });

  describe('Custom Accessors', () => {
    it('should use custom accessor for getting cell values', () => {
      const customColumns = columns.map(col => 
        col.id === 'name' ? {
          ...col,
          accessor: (row: TestData) => row.name.split(' ')[1] // Sort by last name
        } : col
      );

      const sortState: SortState = { columnId: 'name', direction: 'asc' };
      const result = sortEngine.sort(rows, sortState, customColumns);
      
      // Should sort by last name: Brown, Doe, Johnson, Smith, Wilson
      const names = result.map(row => row.data.name);
      expect(names).toEqual(['Alice Brown', 'John Doe', 'Bob Johnson', 'Jane Smith', 'Charlie Wilson']);
    });
  });

  describe('Type Parsing and Comparison', () => {
    it('should parse string numbers for comparison', () => {
      const stringNumberRows = rows.map(row => ({
        ...row,
        data: { ...row.data, age: String(row.data.age) as any }
      }));

      const sortState: SortState = { columnId: 'age', direction: 'asc' };
      const result = sortEngine.sort(stringNumberRows, sortState, columns);
      
      // Should parse strings as numbers and sort correctly
      const ages = result.map(row => parseInt(row.data.age as string));
      expect(ages).toEqual([25, 28, 30, 35, 42]);
    });

    it('should parse currency strings for comparison', () => {
      const currencyRows = rows.map(row => ({
        ...row,
        data: { ...row.data, score: row.data.score ? `$${row.data.score}` as any : null }
      }));

      const sortState: SortState = { columnId: 'score', direction: 'asc' };
      const result = sortEngine.sort(currencyRows, sortState, columns);
      
      // Null first, then parsed currency values
      expect(result[0].data.score).toBeNull();
      
      const scoresWithoutNull = result.slice(1).map(row => 
        parseInt((row.data.score as string).replace('$', ''))
      );
      expect(scoresWithoutNull).toEqual([78, 85, 90, 92]);
    });

    it('should handle invalid date strings gracefully', () => {
      const invalidDateRows = [
        { ...rows[0], data: { ...rows[0].data, joinDate: 'invalid-date' } },
        { ...rows[1], data: { ...rows[1].data, joinDate: '2023-02-20' } },
        { ...rows[2], data: { ...rows[2].data, joinDate: 'another-invalid' } },
      ];

      const sortState: SortState = { columnId: 'joinDate', direction: 'asc' };
      const result = sortEngine.sort(invalidDateRows, sortState, columns);
      
      // Should fall back to string comparison for invalid dates
      const dates = result.map(row => row.data.joinDate);
      expect(dates).toEqual(['2023-02-20', 'another-invalid', 'invalid-date']);
    });
  });

  describe('Performance and Edge Cases', () => {
    it('should handle large datasets efficiently', () => {
      // Create a large dataset
      const largeRows = Array.from({ length: 1000 }, (_, i) => ({
        id: i,
        data: {
          id: i,
          name: `User ${i}`,
          age: 20 + (i % 50),
          email: `user${i}@example.com`,
          active: i % 2 === 0,
          score: i % 100,
          joinDate: `2023-${String((i % 12) + 1).padStart(2, '0')}-01`,
          priority: ['low', 'medium', 'high'][i % 3]
        },
        selected: false,
        expanded: false,
        disabled: false
      }));

      const sortState: SortState = { columnId: 'age', direction: 'asc' };
      const start = performance.now();
      const result = sortEngine.sort(largeRows, sortState, columns);
      const end = performance.now();
      
      expect(result).toHaveLength(1000);
      expect(end - start).toBeLessThan(100); // Should complete in less than 100ms
      
      // Verify it's actually sorted
      const ages = result.map(row => row.data.age);
      for (let i = 1; i < ages.length; i++) {
        expect(ages[i]).toBeGreaterThanOrEqual(ages[i - 1]);
      }
    });

    it('should handle mixed data types gracefully', () => {
      const mixedRows = [
        { ...rows[0], data: { ...rows[0].data, score: 'high' as any } },
        { ...rows[1], data: { ...rows[1].data, score: 85 } },
        { ...rows[2], data: { ...rows[2].data, score: null } },
        { ...rows[3], data: { ...rows[3].data, score: undefined as any } },
        { ...rows[4], data: { ...rows[4].data, score: 90 } },
      ];

      const sortState: SortState = { columnId: 'score', direction: 'asc' };
      const result = sortEngine.sort(mixedRows, sortState, columns);
      
      // Should handle mixed types without crashing
      expect(result).toHaveLength(5);
      
      // Null/undefined should come first, then numbers, then strings
      const scores = result.map(row => row.data.score);
      expect(scores[0]).toBeNull();
      expect(scores[1]).toBeUndefined();
      expect(typeof scores[2]).toBe('number');
      expect(typeof scores[3]).toBe('number');
      expect(typeof scores[4]).toBe('string');
    });

    it('should handle identical values correctly', () => {
      const identicalRows = rows.map(row => ({
        ...row,
        data: { ...row.data, age: 30 } // All same age
      }));

      const sortState: SortState = { columnId: 'age', direction: 'asc' };
      const result = sortEngine.sort(identicalRows, sortState, columns);
      
      // Should maintain stable sort (original order preserved for equal elements)
      expect(result).toHaveLength(5);
      expect(result.every(row => row.data.age === 30)).toBe(true);
    });
  });
});