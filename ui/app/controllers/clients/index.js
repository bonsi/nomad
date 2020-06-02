/* eslint-disable ember/no-incorrect-calls-with-inline-anonymous-functions */
import { alias, readOnly } from '@ember/object/computed';
import { inject as service } from '@ember/service';
import Controller, { inject as controller } from '@ember/controller';
import { computed } from '@ember/object';
import { scheduleOnce } from '@ember/runloop';
import intersection from 'lodash.intersection';
import SortableFactory from 'nomad-ui/mixins/sortable-factory';
import Searchable from 'nomad-ui/mixins/searchable';
import { serialize, deserializedQueryParam as selection } from 'nomad-ui/utils/qp-serialize';

export default Controller.extend(
  SortableFactory(['id', 'name', 'compositeStatus', 'datacenter']),
  Searchable,
  {
    userSettings: service(),
    clientsController: controller('clients'),

    nodes: alias('model.nodes'),
    agents: alias('model.agents'),

    queryParams: {
      currentPage: 'page',
      searchTerm: 'search',
      sortProperty: 'sort',
      sortDescending: 'desc',
      qpClass: 'class',
      qpState: 'state',
      qpDatacenter: 'dc',
      qpVolume: 'volume',
    },

    currentPage: 1,
    pageSize: readOnly('userSettings.pageSize'),

    sortProperty: 'modifyIndex',
    sortDescending: true,

    searchProps: computed(function() {
      return ['id', 'name', 'datacenter'];
    }),

    qpClass: '',
    qpState: '',
    qpDatacenter: '',
    qpVolume: '',

    selectionClass: selection('qpClass'),
    selectionState: selection('qpState'),
    selectionDatacenter: selection('qpDatacenter'),
    selectionVolume: selection('qpVolume'),

    optionsClass: computed('nodes.[]', function() {
      const classes = Array.from(new Set(this.nodes.mapBy('nodeClass')))
        .compact()
        .without('');

      // Remove any invalid node classes from the query param/selection
      scheduleOnce('actions', () => {
        // eslint-disable-next-line ember/no-side-effects
        this.set('qpClass', serialize(intersection(classes, this.selectionClass)));
      });

      return classes.sort().map(dc => ({ key: dc, label: dc }));
    }),

    optionsState: computed(function() {
      return [
        { key: 'initializing', label: 'Initializing' },
        { key: 'ready', label: 'Ready' },
        { key: 'down', label: 'Down' },
        { key: 'ineligible', label: 'Ineligible' },
        { key: 'draining', label: 'Draining' },
      ];
    }),

    optionsDatacenter: computed('nodes.[]', function() {
      const datacenters = Array.from(new Set(this.nodes.mapBy('datacenter'))).compact();

      // Remove any invalid datacenters from the query param/selection
      scheduleOnce('actions', () => {
        // eslint-disable-next-line ember/no-side-effects
        this.set('qpDatacenter', serialize(intersection(datacenters, this.selectionDatacenter)));
      });

      return datacenters.sort().map(dc => ({ key: dc, label: dc }));
    }),

    optionsVolume: computed('nodes.[]', function() {
      const flatten = (acc, val) => acc.concat(val.toArray());

      const allVolumes = this.nodes.mapBy('hostVolumes').reduce(flatten, []);
      const volumes = Array.from(new Set(allVolumes.mapBy('name')));

      scheduleOnce('actions', () => {
        // eslint-disable-next-line ember/no-side-effects
        this.set('qpVolume', serialize(intersection(volumes, this.selectionVolume)));
      });

      return volumes.sort().map(volume => ({ key: volume, label: volume }));
    }),

    filteredNodes: computed(
      'nodes.[]',
      'selectionClass',
      'selectionState',
      'selectionDatacenter',
      'selectionVolume',
      function() {
        const {
          selectionClass: classes,
          selectionState: states,
          selectionDatacenter: datacenters,
          selectionVolume: volumes,
        } = this;

        const onlyIneligible = states.includes('ineligible');
        const onlyDraining = states.includes('draining');

        // states is a composite of node status and other node states
        const statuses = states.without('ineligible').without('draining');

        return this.nodes.filter(node => {
          if (classes.length && !classes.includes(node.get('nodeClass'))) return false;
          if (statuses.length && !statuses.includes(node.get('status'))) return false;
          if (datacenters.length && !datacenters.includes(node.get('datacenter'))) return false;
          if (volumes.length && !node.hostVolumes.find(volume => volumes.includes(volume.name)))
            return false;

          if (onlyIneligible && node.get('isEligible')) return false;
          if (onlyDraining && !node.get('isDraining')) return false;

          return true;
        });
      }
    ),

    listToSort: alias('filteredNodes'),
    listToSearch: alias('listSorted'),
    sortedNodes: alias('listSearched'),

    isForbidden: alias('clientsController.isForbidden'),

    setFacetQueryParam(queryParam, selection) {
      this.set(queryParam, serialize(selection));
    },

    actions: {
      gotoNode(node) {
        this.transitionToRoute('clients.client', node);
      },
    },
  }
);
