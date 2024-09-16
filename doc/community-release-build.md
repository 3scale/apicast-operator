# APIcast-Operator Community Release

[Introduction](#introduction)  
[Main Steps Briefly](#main-steps-briefly)  
[Versioning Strategy](#versioning-strategy)  
[Steps to Perform Community Release in apicast-operator repo](#steps-to-perform-community-release-in-3scale-operator-repo)  
[Prepare release in community-operators-prod repo](#prepare-release-in-community-operators-prod-repo)
[References](#references)


## Introduction
* Community Release is primarily created in parallel with minor or patch Product Releases, but a major community may also be released.
  * Community Release is not required for every productized build or every product patch release.
* Community release can be released after downstream release
* Community Release codebase is defined by Tag in the [apicast-operator](https://github.com/3scale/apicast-operator) repository. 
Tag format as v\<major>.\<minor>.<build#> (for example: v0.6.0, v0.7.1)
* Community Release bundles are defined in `redhat-openshift-ecosystem` in [apicast-community-operator repo](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/apicast-community-operator) . 


## Main Steps Briefly

Below are briefly the main steps for prepare Community Release

### APIcast
* Tagging - Search [APIcast repository](https://github.com/3scale/APIcast) for APIcast version and Tag to be used 
in Community APIcast operator. For example: tag: v3.15.0 (release 3.15.0)


### APIcast operator

* Create a community release development branch in **apicast-operator** repo, based on the Product release CommitID, 
  or, if Product release tag is not yet available, - based on **3scale-2.X-stable** branch. Do development.
* Create PR from community release development branch to a **3scale-2.X-stable** branch (PR example: https://github.com/3scale/apicast-operator/pull/226)
  * Do development, Test and Merge PR to `3scale-2.X-stable` branch
    * **Note**: Build of the PR will create a release image, like https://quay.io/repository/3scale/apicast-operator?tab=tags&tag=v0.7.1
* Testing - Following Testing will be done:
  * Initial installation from Index image,
  * Upgrade from previous release
* Tagging - Apply community release Tag (like v0.8.0) to the latest CommitId on the community release development branch.
* Prepare Community release - create PR in **community-operators-prod** repo (based on manifests from **apicast-operator** repo), 
as in [PR example - apicast-community-operator (0.7.1)](https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/2382)
  * **Merge of this PR will publish Release**.
* Do sanity test of the Release after publishing

## Versioning Strategy
Semantic Versioning scheme will be used for Community releases:  MAJOR.MINOR.PATCH.
* MAJOR version: Increments when breaking changes are introduced
* MINOR version: Increments when backward-compatible functionality is added.
* PATCH version: Increments when backward-compatible bug fixes are made.
* APIcast ooperator Community releases Versions examples:  0.6.0,  0.7.1,  0.8.0
* The Version is used in Community releases Tag name

```shell
$ git tag |grep -E "0.6|0.7"     
v0.6.0
v0.7.0
v0.7.1
```

* Tag should be signed. Example of signed tag:
```  
 git show v0.7.1
tag v0.7.1
Tagger: Eguzki Astiz Lezaun <eastizle@redhat.com>
Date:   Tue Mar 14 10:46:34 2023 +0100

v0.7.1
-----BEGIN PGP SIGNATURE-----
xxxxx
xxxxx
-----END PGP SIGNATURE-----

commit 4f3a2e26fa4e97f28cd0d276dc5d46261354afd4 (tag: v0.7.1, tag: show, tag: 3scale-2.13.2-GA, tag: 3scale-2.13.1-GA)
.....
```

## Steps to Perform Community Release in 3scale-operator repo
1. Community Release Development Branch

* Create a Community release development branch **based on Product Release CommitID**
* Create PR and target it to **3scale-2.X-stable** branch

```
$ cd apicast-operator
$ git checkout -b apicast-community-v0.8.0 f76e0a60c8590ce8d463180c1f6d558c56d3d9c2
```
Note. Commit ID in this example was found as latest commit id from `git log origin/3scale-2.15-stable` 

2. APIcast Components Images - APIcast
* Community release can use its own APIcast image or use existing build of APIcast. 
  You can search [APIcast repository](https://github.com/3scale/APIcast) for APIcast version and Tag to be used
  in Community APIcast operator. For example: tag: v3.15.0 (release 3.15.0)


3. Create PR in apicast-operator upstream for new bundle
   * Work to have E2E tests passed in PR
   * Merge PR to **3scale-2.X-stable** branch
  
4. Testing
   
   * Testing of `apicast-operator` must be completed before open PR in `community-operators-prod` repo
   * Testing must be completed before Tagging
   * Testing must be done for `Fresh Install` and `Upgrade`
  

5. Tagging

- Community Release is defined by Tag.  
- Tag will be applied to the Community Release Branch
- Tag will be applied after Testing completion.
- Tag will be signed

```
$ cd apicast-operator
$ git tag -s v0.8.0 -m "APIcast operator Community Release v0.8.0 , based on <commit id>"
$ git push myfork v0.11.0
```
**Note**. Signed tags use GPG (GNU Privacy Guard) to create a cryptographic signature (GPG installation/details are not provided in this doc)

## Prepare release in community-operators-prod repo
See Documentation in [Pull Requests in community operators project](https://github.com/operator-framework/community-operators/blob/master/docs/contributing-via-pr.md)

**IMPORTANT. All previouse steps, including Testing of apicast-operator (Fresh install and Upgrade) must be completed before opening a PR in the community-operators-prod repo**

* Fork & Clone community-operators-prod repo if you don't have it yet
* Get latest changes

```
$ git clone git@github.com:redhat-openshift-ecosystem/community-operators-prod.git
$ cd community-operators-prod

Example for myfork:
$ git remote -v
myfork        git@github.com:valerymo/community-operators-prod.git (fetch)
myfork        git@github.com:valerymo/community-operators-prod.git (push)
origin        git@github.com:redhat-openshift-ecosystem/community-operators-prod.git (fetch)
origin        git@github.com:redhat-openshift-ecosystem/community-operators-prod.git (push)
```

* Create branch in community-operators-prod repo
```
$ git checkout -b apicast-community-operator-v0.8.0
```

* Create a new release folder (similar to previous release)
```
community-operators-prod/operators/apicast-community-operator/0.8.0
```

* Copy manifests from apicast-operator Community release to community-operators-prod, as for example From  [apicast-operator PR](https://github.com/3scale/apicast-operator/pull/226) To [community-operators-prod/operators/apicast-community-operator/0.8.0](community-operators-prod/operators/apicast-community-operator/0.8.0)

* Update CSV. These are things that need pay attention:
  *  apicast-operator containerImage
  *  apicast image
  *  CVS version
  *  CSV replaces
  *  CSV name
  *  CSV description
  *  CSV urls
  *  Annotations
  * * Update metadata/annotations  and bundle.Dockerfile
  

* Compare apicast-community-operator bundle with previous version, and with apicast-operator
* Finally you will have update release bundle structure.
* Commit your changes and Create PR 
  * Do Signed-off  - git commit -s … , It’s required for PR test Pipeline

```
[community-operators-prod] (apicast-community-operator-v0.8.0)$ git log
commit xxxxxxxxxxxx (HEAD -> apicast-community-operator-v0.8.0)
Author: xxx xxx <xxx@xxx.com>
Date:   xxx
apicast-community-operator release v0.8.0
Signed-off-by: xxx xxx <xxx@xxx.com>
```

```
$ git push -u myfork apicast-community-operator-v0.8.0
To github.com:valerymo/community-operators-prod.git
....
```

**IMPORTANT. Merging of community-operators-prod PR will trigger Release creation/publishing**

* Check and confirm all questions in PR description
* Merge
* Test of created release in OSD cluster / OperatorHub
________________


## References

* [Community operators project documentation](https://github.com/operator-framework/community-operators/blob/master/docs/contributing-via-pr.md)
* [Community operators repository](https://github.com/k8s-operatorhub/community-operators)
* [K8S PR example](https://github.com/k8s-operatorhub/community-operators/pull/2472)
* [OCP PR example](https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/2382)
* *Check for the latest guidelines about what it needs to be done; they change from time to time*
  * *Community operator release process moved* [here - operator-release-process](https://github.com/operator-framework/community-operators/blob/master/docs/operator-release-process.md)
