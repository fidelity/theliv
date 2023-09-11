/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs';
import { ActivatedRoute, Router } from '@angular/router';
import { faSearch, faTimes, faSpinner, faBell, faCheck, faPencilAlt, faAngleDown, faAngleUp, faVideo} from '@fortawesome/free-solid-svg-icons';
import { KubernetesService } from '../services/kubernetes.service';
import { map, startWith } from 'rxjs/operators';
import { FormControl } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { UserFeedbackComponent } from '../components/user-feedback/user-feedback.component';

export interface NamespaceOption {
  text: string
  value: string
}

@Component({
  selector: 'app-kube-platform',
  templateUrl: './kube-platform.component.html',
  styleUrls: ['./kube-platform.component.scss']
})
export class KubePlatformComponent implements OnInit {
  faSearch = faSearch;
  faTimes = faTimes;
  faSpinner = faSpinner;
  faBell = faBell;
  faCheck = faCheck;
  faPencialAlt = faPencilAlt;
  faAngleDown = faAngleDown;
  faAngleUp = faAngleUp;
  faVideo = faVideo;
  loading = false;
  public resourceTypes: any;
  public proDomains: any;
  public resourceNames: any;
  public senLevels: any;
  public resourceStatus: any;
  resourceGroups: any;
  allResources: any;
  typeFilter: any[] = [];
  nameFilter: any[] = [];
  domainFilter: any[] = [];
  isGroupContent = false;
  count = -1;
  sortBy = 'time';
  now = new Date();

  type: any;
  selectedNs = '';
  selectedClusters = '';
  selectedType = '';
  selectedName = '';
  //namespaces: any;
  clusters: any;

  clusterFormControl = new FormControl();
  clusterOptions: Observable<string[]> | undefined;
  namespaces: NamespaceOption[] = []
  clusterInputing = false

  events: any;
  hasFailedEvents=false;
  sortedData: any;
  isAsc=false;
  active='time';
  feedback = '';
  gridToggle = false;
  configInfo: any;

  constructor(
    private kubeService: KubernetesService,
    private route: ActivatedRoute,
    private router: Router, 
    private dialog: MatDialog) { }

  ngOnInit(): void {
    this.resourceGroups = null;
    // this.router.routeReuseStrategy.shouldReuseRoute = () => false;
    this.kubeService.getClusters().subscribe((res: any) => {
      if (res) {
        this.clusters = res;
        this.clusterOptions = this.clusterFormControl.valueChanges.pipe(
          startWith(''),
          map(value => {
            const filterValue = value.toLowerCase();
            return this.clusters.filter((option: string) =>  
              option.toLowerCase().indexOf(filterValue) >= 0 
            );
          })
        );
      }
    }, (err: any) => {
      console.log('Get Clusters Information Error: ', err);
    });

    this.route.queryParamMap.subscribe((params: any) => {
      this.resourceGroups = [];
      if (params.get('cluster')) {
        this.selectedClusters = params.get('cluster')
      }
      this.kubeService.selectedClusters$.next(this.selectedClusters);

      if (params.get('namespace')) {
        this.selectedNs = params.get('namespace');
      }
      if (params.get('type')) {
        this.selectedType = params.get('type').charAt(0).toUpperCase() + params.get('type').slice(1);
      }
      if (params.get('name')) {
        this.selectedName = params.get('name');
      }
      this.kubeService.selectedNs$.next(this.selectedNs);

      this.getNamespaces()
      if (this.selectedClusters && this.selectedNs) {
        this.getKubeResourceInfo();
        this.getEvents();
      }
    });

    this.kubeService.getConfigInfo().subscribe((res: any) => {
      if (res) {
        this.configInfo = res;
      }
    }, (err: any) => {
      console.log('Get Config Information Error: ', err);
    });
  }

  getKubeResourceInfo(): void {
    this.loading = true;
    this.kubeService.getDetects(this.selectedClusters, this.selectedNs).subscribe(
      (res: any) => {
        if (res) {
          this.resourceGroups = res;
          this.allResources = res;
          this.kubeService.resourceList$.next(this.resourceGroups);
          this.isGroupContent = true;
          this.count = res.length;
          this.loading = false;
          this.getProblemNameFilter()
          this.getResourceTypeFilter()
          this.getProblemDomainFilter()
          this.filterByQuery()
        }
      },
      (err: any) => {
        console.log('Get Kube Information Error: ', err);
      }
    );
  }

  getResourceTypeFilter(): void {
    var list: any[] = [];
    this.allResources.forEach((r: any) => {
      if (!list.find((item: any) => item.name === r.topResourceType)) {
        var n = r.topResourceType;
        var c = 1;
        var check = false;
        if (n == this.selectedType) {
          check = true;
          this.typeFilter.push(n)
        }
        var resource = {name: n, count: c, isChecked: check}
        list.push(resource)
      } else{
        var obj = list.find((item: any) => item.name === r.topResourceType)
        obj.count = obj.count + 1;
      }
    });
    var totalCount = 0;
    list.forEach((item: any)=>{
      totalCount = totalCount+ item.count;
    })
    list.push({name: 'All Types', count: totalCount})
    console.log(list);
    this.resourceTypes = list;
  }

  getProblemDomainFilter(): void {
    var list: any[] = [];
    this.allResources.forEach((r: any) => {
      if (!list.find((item: any) => item.name === r.rootCause.name)) {
        var n = r.rootCause.name;
        var c = 1;
        var resource = {name: n, count: c}
        list.push(resource)
      } else{
        var obj = list.find((item: any) => item.name === r.rootCause.name)
        obj.count = obj.count + 1;
      }
    });
    var totalCount = 0;
    list.forEach((item: any)=>{
      totalCount = totalCount+ item.count;
    })
    list.push({name: 'All Domains', count: totalCount })
    console.log(list);
    this.proDomains = list;
  }

  getProblemNameFilter(): void {
    var list: any[] = [];
    this.allResources.forEach((r: any) => {
      if (!list.find((item: any) => item.name === r.name)) {
        var n = r.name;
        var c = 1;
        var check = false;
        if (n == this.selectedName) {
          this.nameFilter.push(n);
          check = true;
        }
        var resource = {name: n, count: c, isChecked: check}
        list.push(resource)
      } else{
        var obj = list.find((item: any) => item.name === r.name)
        obj.count = obj.count + 1;
      }
    });
    var totalCount = 0;
    list.forEach((item: any)=>{
      totalCount = totalCount+ item.count;
    })
    list.push({name: 'All', count: totalCount })
    this.resourceNames = list;
  }

  filterByQuery(): void {
    if (this.selectedType) {
      this.resourceTypes.forEach((r: any) => {
        if (r.name == this.selectedType) {
          this.resourceGroups = this.resourceGroups.filter((r: any) => r.topResourceType == this.selectedType)
        }
      })
    }
    if (this.selectedName) {
      this.resourceNames.forEach((r: any) => {
        if (r.name == this.selectedName) {
          this.resourceGroups = this.resourceGroups.filter((r: any) => r.name == this.selectedName)
        }
      })
    }
    this.kubeService.resourceList$.next(this.resourceGroups);
  }

  filter: any[] = [];
  changeSelection(e: any, cat: string): void {
    if (e.target.checked && e.target.value) {
      console.log('checked: ', e.target.value);
      switch (cat) {
        case 'type':
          this.typeFilter.push(e.target.value);
          break;
        case 'domain':
          this.domainFilter.push(e.target.value);
          break;
        case 'name':
          this.nameFilter.push(e.target.value);
          break;
        default:
          break;
      }
    } else if (!e.target.checked && e.target.value) {
      switch (cat) {
        case 'type':
          this.typeFilter.splice(this.typeFilter.findIndex((element) => element == e.target.value), 1);
          break;
        case 'domain':
          this.domainFilter.splice(this.domainFilter.findIndex((element) => element == e.target.value), 1);
          break;
        case 'name':
          this.nameFilter.splice(this.nameFilter.findIndex((element) => element == e.target.value), 1);
          break;
        default:
          break;
        }
    }
    this.filterResource();
    this.kubeService.resourceList$.next(this.resourceGroups);
    // else if (this.type.indexOf(e.target.value) > -1) {
    //   const index = this.type.indexOf(e.target.value, 0);
    //   this.type.splice(index, 1);
    // }
  }

  filterResource(): void {
    var resultList: any;
    resultList=this.allResources;
    if (this.nameFilter.length > 0 && !this.nameFilter.includes("All")) {
      resultList=resultList.filter((r: any) => this.nameFilter.includes(r.name));
    }
    if (this.typeFilter.length > 0 && !this.typeFilter.includes("All Types")) {
      resultList=resultList.filter((r: any) =>  this.typeFilter.includes(r.topResourceType))
    }
    if (this.domainFilter.length > 0 && !this.domainFilter.includes("All Domains")) {
      resultList=resultList.filter((r: any) =>  this.domainFilter.includes(r.rootCause.name))
    }
    this.resourceGroups = resultList;
  }

  checkClusterBlank(){
    this.clusterInputing = true
    if (this.selectedClusters == '') {
      this.selectedNs = ''
      this.selectedType = ''
      this.selectedName = ''
      this.namespaces = []
      this.resourceGroups = []
      this.resourceTypes = []
      this.proDomains = []
      this.resourceNames = []
      this.typeFilter = []
      this.nameFilter = []
      this.domainFilter = []
      this.events = []
    }
  }

  getNSByCluster(): void {
    this.clusterInputing = false
    this.router.navigate(['kubernetes'], { queryParams: { cluster: this.selectedClusters } });
    // this.resourceGroups = [];
    // this.selectedNs = '';
    // this.kubeService.selectedNs$.next(this.selectedNs);
    // this.kubeService.resourceList$.next(this.resourceGroups);
    // this.getNamespaces()
  }

  getNamespaces(): void {
    if (this.selectedClusters) {
      this.kubeService.getAllNamespaces(this.selectedClusters).subscribe(
        (res: any) => {
          if (res) {
            this.namespaces = []
            res.forEach((ele: string) => {
              this.namespaces.push({text: ele, value: ele})
            });
          }
        },
        (err: any) => {
          console.log(`Get Namespace in Cluster ${this.selectedClusters} Error: `, err);
        }
      );
    }
  }

  getSelectedQuery(e: any): void {
    this.router.navigate(['kubernetes'], { queryParams: { cluster: this.selectedClusters, namespace: e.value } });
    // this.resourceGroups = [];
    // this.kubeService.selectedClusters$.next(this.selectedClusters);
    // this.kubeService.selectedNs$.next(this.selectedNs);
    // if (this.selectedClusters && this.selectedNs) {
    //   this.getKubeResourceInfo();
    // }
  }


  openFeedbackDialog(): void {
    const dialogRef = this.dialog.open(UserFeedbackComponent, {
      data: {feedback: this.feedback},
    });

    dialogRef.afterClosed().subscribe(result => {
      result = 'The dialog was closed';
      console.log(result);
      this.feedback = '';
    });
  }

  getEvents(): void {
    this.hasFailedEvents=false;
    this.kubeService.getKubeEvents(this.selectedClusters, this.selectedNs).subscribe((res: any) => {
      if (res) {
        this.events = res;
        this.events.sort((a:any, b:any) => {
          return compare(a.DateHappened, b.DateHappened, false);
        });
        if (this.events.find((e: any) => e.Type!=="Normal")) {
          this.hasFailedEvents = true;
        }
      }
    }, (err: any) => {
      console.log('Get Kube Events Error: ', err);
    });
  }

  showEvents(): void {
    this.gridToggle=true;
  }

  close(): void {
    this.gridToggle = false;
  }

  sortData(header: string) {
    this.events.sort((a:any, b:any) => {
      var isAsc = this.isAsc;
      switch (header) {
        case 'type':
          return compare(a.Type, b.Type, isAsc);
        case 'resource':
          return compare(a.InvolvedObject.name, b.InvolvedObject.name, isAsc);
        case 'kind':
          return compare(a.InvolvedObject.Kind, b.InvolvedObject.Kind, isAsc);
        case 'reason':
          return compare(a.Reason, b.Reason, isAsc);
        case 'time':
          return compare(a.DateHappened, b.DateHappened, isAsc);
        default:
          return 0;
      }
    });
  }

}

function compare(a: number | string, b: number | string, isAsc: boolean) {
  return (a < b ? -1 : 1) * (isAsc ? 1 : -1);
}